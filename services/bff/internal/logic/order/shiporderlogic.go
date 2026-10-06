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

type ShipOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShipOrder 订单发货(管理端)。:id 是 order_id。
//
// operator 从 JWT 的 username 取(proto 的字段名是 user_name,
// 与 CancelOrder 的 operator 不同)。
//
// 响应是单体 response.OrderShipResp 的五个字段,比 proto 的
// ShipOrderResp 多 order_no(那个要从内层 Order 取)。
func (l *ShipOrderLogic) ShipOrder(req *types.ShipOrderReq) (*types.ShipOrderResp, error) {
	userName, _ := middleware.Username(l.ctx)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.ShipOrder(ctx, &v1_tradev1.ShipOrderReq{
		OrderId:        req.Id,
		Username:       userName,
		ExpressCompany: req.ExpressCompany,
		TrackingNo:     req.TrackingNo,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 / 状态不是"待发货" → 400
		return nil, err
	}

	// proto 的 ShipOrderResp 是 {Order order, express_company, tracking_no}
	// —— order_id / order_no / order_status 都在内层 Order 里,
	// 而单体的 response.OrderShipResp 把它们摊平到外层。
	o := resp.GetOrder()
	return &types.ShipOrderResp{
		OrderId:        o.GetOrderId(),
		OrderNo:        o.GetOrderNo(),
		OrderStatus:    o.GetOrderStatus(),
		ExpressCompany: resp.GetExpressCompany(),
		TrackingNo:     resp.GetTrackingNo(),
	}, nil
}
