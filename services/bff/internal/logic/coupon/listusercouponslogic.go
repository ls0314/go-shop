// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package coupon

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserCouponsLogic {
	return &ListUserCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListUserCoupons 分页查询当前用户持有的券("我的卡券")。
//
// ============================================================
// user_id 从 JWT 取,HTTP 请求里没有这个参数
// ============================================================
//
// proto 的 ListUserCouponsReq **必须**带 user_id(服务端按它过滤),
// 而 .api 里 UserCouponListReq 只有 {page, page_size, status}。
// 这是刻意的:券的归属必须由服务端从令牌决定,否则任何登录用户都能
// 查别人的券包。
//
// 取不到身份返回 errNoIdentity → 500。走到那一格说明**该路由没挂
// Auth 中间件**(配置错误),不是用户的问题,故不能静默当成"没有券"
// 返回空列表 —— 那会让配置漏挂一直不被发现。
//
// ============================================================
// status 直传,不校验也不转换
// ============================================================
//
// 取值 unused / used / expired,空串表示不过滤(proto 注释)。
// BFF 不校验:传错了会得到空列表而不是报错,那是服务端的行为;
// 要在这里拦就得维护一份状态枚举,而那份枚举必然与服务端漂移。
//
// 这里与"结算可用券"的口径不同:本接口**不过滤过期**(服务端 repo
// 的注释:"我的卡券页要展示已用/已过期"),故 status 传空时三种状态
// 的券都会返回 —— 前端要按 status 自己分 tab。
//
// BFF **也不拿 expire_at 自己判过期**:过期迁移是券域对账任务的活
// (couponreconcile 把 unused 且已过期的行置为 expired),两端各自判
// 时间会给出不同结论 —— 与模板 status 不重算是同一条理由。
//
// ============================================================
// 分页兜底并回写
// ============================================================
//
// 与 order 域同口径:不传 → 1/10,page_size 上限 100,兜底后的值回写
// 到响应供前端渲染分页控件。本域是 page_size(snake_case)。
func (l *ListUserCouponsLogic) ListUserCoupons(req *types.UserCouponListReq) (*types.UserCouponListResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.ListUserCoupons(ctx, &v1_marketingv1.ListUserCouponsReq{
		UserId:   userId,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   req.Status,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.UserCouponListResp{
		List:     toUserCouponItems(resp.GetItems()),
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
