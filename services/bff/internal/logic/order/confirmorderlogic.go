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

type ConfirmOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmOrderLogic {
	return &ConfirmOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ConfirmOrder 确认收货(用户端)。
//
// ============================================================
// 与 CancelOrder 的对称与不对称
// ============================================================
//
// 对称之处:都是**用户对自己订单的状态流转**,proto 入参都是
// {order_id, user_id, user_name},服务端查
// `WHERE order_id = ? AND user_id = ?` 保证归属。
//
// 不对称之处:
//
//	CancelOrder  要**逆序补偿**(释放库存 + 退还券),
//	             故多一个 compensated 字段供上层告警
//	ConfirmOrder 是**纯状态流转**(待收货 → 已完成),无需补偿,
//	             故响应只有 {order, error_msg}
//
// 这个不对称是合理的:取消要退还资源,确认收货只是把"货已到手"
// 记下来 —— 它触发的是**结算方向的后续**(如给商家放款),
// 而那不在 trade 的事务里(DS-A-26 的对账任务负责)。
//
// ============================================================
// user_name 的用途
// ============================================================
//
// proto 的 ConfirmOrderReq.user_name —— 与 CancelOrder 的 operator
// 一样,用于订单日志的"操作人"。
//
// **注意字段名不同**:CancelOrder 叫 operator,ConfirmOrder 叫
// user_name。proto 自己就不统一。BFF 两处都传 JWT 的 username。
//
// ============================================================
// 响应
// ============================================================
//
// 单体 response.OrderStatusResp 是 {order_id, order_no, order_status}。
//
// proto 的 ConfirmOrderResp 是 {Order order, error_msg} ——
// order 是内层消息,单体把它摊平到外层。故这里取三个字段。
func (l *ConfirmOrderLogic) ConfirmOrder(req *types.OrderIdReq) (*types.ConfirmOrderResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}
	userName, _ := middleware.Username(l.ctx)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.ConfirmOrder(ctx, &v1_tradev1.ConfirmOrderReq{
		OrderId:  req.Id,
		UserId:   userId,
		Username: userName,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 / 不属于当前用户 / 状态不是"待收货" → 400
		return nil, err
	}

	o := resp.GetOrder()
	return &types.ConfirmOrderResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
	}, nil
}
