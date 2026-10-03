package couponclient

import (
	"context"
	"errors"
	"time"

	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrUnavailable 客户端未建连(etcd 连不上 / 服务未注册)时的统一错误。
//
// 与 productclient.ErrUnavailable 同构:文案本身是契约 ——
// HTTP 层据此回 503,接线冒烟测试据此跳过该路由。
var ErrUnavailable = errors.New("marketing-service 不可用")

// RestoreError 把服务端返回的 error_msg 还原成本地哨兵错误(能还原时)或普通错误。
//
// 为什么必须还原而不是 errors.New(msg) 了事:
//   - **依赖缺失要能被识别**:HTTP 层靠 errors.Is(err, ErrUnavailable) 决定回 503
//     还是 500,丢了哨兵就只能一律 500;
//   - **业务失败要能被分类**:metrics 埋点靠 errors.Is 把"售罄/超限"与
//     "其它错误"分开计数(见 service/coupon_metrics.go),分类塌成一种,
//     既有的告警与看板立刻失真。
//
// 还原失败的按普通错误返回 —— 调用方仍可拿到文案,只是无法 errors.Is。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if err, ok := serverErrMap[errorMsg]; ok {
		return err
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// serverErrMap 服务端文案 → 本地哨兵错误。
//
// 绝大多数文案两侧逐字一致,靠 error.Error() 比对即可;这里只登记
// **两侧不一致**的两条:单体原文案把"券"写成了"卷"(错别字),
// marketing-service 迁出时改正了,于是出现同义不同文:
//
//	服务端 "无权使用该优惠券"  ←→  本地 model.ErrUseCouponNoNoPermission ("…优惠卷")
//	服务端 "返还优惠券失败"    ←→  本地 model.ErrCannotCancelCoupon   ("返还优惠卷失败")
//
// 映射放在这里而不是去改单体的错别字:那个文案可能已经被别处引用,
// 改名要单独评估;而映射表只有两行,且天然随时间收敛(等 C4 把订单链路
// 也改成 RPC 后,这两条的错误处理会一并重做)。
var serverErrMap = map[string]error{
	"无权使用该优惠券": model.ErrUseCouponNoNoPermission,
	"返还优惠券失败":  model.ErrCannotCancelCoupon,
}

// couponCallTimeout 单次券 RPC 的超时。
// 比 productclient 的 3s 略长:领券是写操作且服务端要拿行锁排队,
// 超时太短会把"正在排队"误报成失败。
const couponCallTimeout = 5 * time.Second

// CouponClient 券域(marketing-service)的 RPC 客户端。
//
// 覆盖该服务对外暴露的全部 11 个方法:管理端模板 CRUD、用户端券中心、
// 以及供下单/取消调用的 UseCoupon / ReturnCoupon。
type CouponClient struct {
	coupon v1_marketingv1.CouponServiceClient
	conn   *grpc.ClientConn
}

// NewCouponClient 建连 marketing-service(etcd 服务发现)。
func NewCouponClient(etcdHosts []string, etcdKey string) (*CouponClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	return &CouponClient{
		coupon: v1_marketingv1.NewCouponServiceClient(client.Conn()),
		conn:   client.Conn(),
	}, nil
}

func (c *CouponClient) Close() error { return c.conn.Close() }

// ============================================================
// 管理端:模板
// ============================================================

// CreateCouponTemplate 创建券模板。返回 (templateId, errMsg, err)。
// errMsg 非空表示业务失败(参数非法 / 有效期配置非法),由调用方决定 HTTP 码。
func (c *CouponClient) CreateCouponTemplate(req requset.CreateCouponReq) (int64, string, error) {
	if c == nil {
		return 0, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	protoReq := &v1_marketingv1.CreateCouponTemplateReq{
		Template: &v1_marketingv1.CouponTemplate{
			CouponName:      req.CouponName,
			CouponType:      req.CouponType,
			ThresholdAmount: req.ThresholdAmount,
			DiscountAmount:  req.DiscountAmount,
			TotalCount:      req.TotalCount,
			PerUserLimit:    req.PerUserLimit,
			UsableDays:      req.UsableDays,
		},
	}
	// 固定有效期:前端未传时为零值时间,转成 nil 让服务端按"未配置"处理,
	// 否则 0001-01-01 会被当成一个合法的过去时间,有效期校验就走了另一条分支
	if !req.StartTime.IsZero() {
		protoReq.Template.StartTime = timestamppb.New(req.StartTime)
	}
	if !req.EndTime.IsZero() {
		protoReq.Template.EndTime = timestamppb.New(req.EndTime)
	}

	resp, err := c.coupon.CreateCouponTemplate(ctx, protoReq)
	if err != nil {
		return 0, "", err
	}
	if resp.ErrorMsg != "" {
		return 0, resp.ErrorMsg, nil
	}
	return resp.Template.GetTemplateId(), "", nil
}

// GetCouponList 管理端模板分页列表
func (c *CouponClient) GetCouponList(req requset.GetCouponListReq) (*response.GetCouponListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ListCouponTemplates(ctx, &v1_marketingv1.ListCouponTemplatesReq{
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		CouponName: req.CouponName,
		CouponType: req.CouponType,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelCouponList(resp), "", nil
}

// ============================================================
// 用户端:我的券 / 领券中心 / 结算可用券
// ============================================================

// GetUserCouponList 我的卡券分页列表(不过滤过期,已用/已过期也展示)
func (c *CouponClient) GetUserCouponList(userId int64, req requset.UserGetCouponListReq) (*response.UserGetCouponListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ListUserCoupons(ctx, &v1_marketingv1.ListUserCouponsReq{
		UserId:   userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Status:   req.Status,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelUserCouponList(resp), "", nil
}

// GetReceiveCouponList 领券中心:可领模板 + 该用户已领数/剩余量
func (c *CouponClient) GetReceiveCouponList(userId int64, req requset.UserGetTemplateListReq) (*response.UserCouponTemplateListResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ListCouponTemplatesForUser(ctx, &v1_marketingv1.ListCouponTemplatesForUserReq{
		UserId:   userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelTemplateList(resp), "", nil
}

// GetAvailableCouponList 结算可用券(order_amount 用于门槛判断与实付计算)
func (c *CouponClient) GetAvailableCouponList(userId int64, req requset.GetAvailableCouponReq) (*response.GetAvailableCouponResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ListAvailableCoupons(ctx, &v1_marketingv1.ListAvailableCouponsReq{
		UserId:      userId,
		OrderAmount: req.OrderAmount,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelAvailableList(resp), "", nil
}

// ReceiveCoupon 领券。返回 (userCouponId, expireAt, errMsg, err)。
func (c *CouponClient) ReceiveCoupon(userId, templateId int64) (*response.UserReceiveCouponResp, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ReceiveCoupon(ctx, &v1_marketingv1.ReceiveCouponReq{
		UserId:     userId,
		TemplateId: templateId,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}

	out := &response.UserReceiveCouponResp{UserCouponId: resp.UserCouponId}
	if resp.ExpireAt != nil {
		out.ExpireTime = resp.ExpireAt.AsTime()
	}
	return out, "", nil
}

// ============================================================
// 供下单/取消调用:核销 / 归还
// ============================================================

// UseCouponResult 核销结果:券快照 + 服务端算好的实付金额
type UseCouponResult struct {
	Coupon    *model.UserCoupon
	PayAmount float64
}

// UseCoupon 下单核销券。
//
// 幂等键是 orderNo:同一订单重复调用只生效一次。
// 返回 (result, errMsg, err):errMsg 非空为业务失败(不存在/已用/越权/未达门槛)。
func (c *CouponClient) UseCoupon(userCouponId int64, orderNo string, userId int64, orderAmount float64) (*UseCouponResult, string, error) {
	if c == nil {
		return nil, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.UseCoupon(ctx, &v1_marketingv1.UseCouponReq{
		UserCouponId: userCouponId,
		OrderNo:      orderNo,
		UserId:       userId,
		OrderAmount:  orderAmount,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return &UseCouponResult{
		Coupon:    toModelUserCoupon(resp.Coupon),
		PayAmount: resp.PayAmount,
	}, "", nil
}

// ReturnCoupon 取消订单归还券(按订单号反查,幂等)。
//
// 返回 (returned, errMsg, err):returned=false 且无错误表示
// "该订单没用券"或"券已归还" —— 两种情况都算补偿成功。
func (c *CouponClient) ReturnCoupon(orderNo string, userId int64) (bool, string, error) {
	if c == nil {
		return false, "", ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), couponCallTimeout)
	defer cancel()

	resp, err := c.coupon.ReturnCoupon(ctx, &v1_marketingv1.ReturnCouponReq{
		OrderNo: orderNo,
		UserId:  userId,
	})
	if err != nil {
		return false, "", err
	}
	if resp.ErrorMsg != "" {
		return false, resp.ErrorMsg, nil
	}
	return resp.Returned, "", nil
}
