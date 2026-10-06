// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOrderLogic {
	return &GetUserOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetUserOrder 取单个订单的详情(用户端)。
//
// ============================================================
// 归属校验在服务端
// ============================================================
//
// proto 的 GetUserOrderReq 是 {order_id, user_id},服务端的查询带着
// `WHERE order_id = ? AND user_id = ?` —— 查别人的订单会查不到。
//
// 故 BFF 只传 JWT 的 userId,不预检。
//
// 这与同组档案域(/admin/user/info/:id)不同 —— 那里的 proto 只收
// 一个 user_id,服务端无法判断请求者身份,故 BFF 必须比对。
//
// ============================================================
// 三处与 proto 不同的映射
// ============================================================
//
//	① details → **detail_list**;logs → **log_list**(改名)
//	② address_snapshot 从**字符串**解析成**对象**
//	③ proto 的 order 内层消息(GetUserOrderResp 里有个 Order)
//	   它带 user_id / expire_at / idempotent_key —— 单体都不返回
//
// 三者都在 convert.go 里说明过。
//
// ============================================================
// 空列表的语义
// ============================================================
//
// 一个刚下单但还没产生日志的订单,log_list 是空数组([])而不是 null
// —— toOrderLogs 用 make(..., 0, len)。前端 map/length 不会抛错。
func (l *GetUserOrderLogic) GetUserOrder(req *types.OrderIdReq) (*types.GetUserOrderResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.GetUserOrder(ctx, &v1_tradev1.GetUserOrderReq{
		OrderId: req.Id,
		UserId:  userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 / 不属于当前用户 → 400
		return nil, err
	}

	// GetUserOrderResp 里有个 Order 内层消息,单体的
	// response.UserGetOrderResp 是把它的字段**摊平**到外层的
	// (order_id / order_no / order_status / ... 直接在外层)。
	//
	// 故这里也要摊平,而不是把 Order 整个放进去。
	o := resp.GetOrder()
	return &types.GetUserOrderResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
		TotalAmount: o.GetTotalAmount(),
		PayAmount:   o.GetPayAmount(),
		PayMethod:   o.GetPayMethod(),
		PayTime:     formatTimestamp(o.GetPayTime()),
		// 字符串 → 对象(见 convert.go 的 parseSnapshot)
		AddressSnapshot: parseSnapshot(o.GetAddressSnapshot()),
		BuyerRemark:     o.GetBuyerRemark(),
		// details → detail_list;logs → log_list
		DetailList: toOrderDetails(resp.GetDetails()),
		LogList:    toOrderLogs(resp.GetLogs()),
		CreatedAt:  formatTimestamp(o.GetCreatedAt()),
	}, nil
}
