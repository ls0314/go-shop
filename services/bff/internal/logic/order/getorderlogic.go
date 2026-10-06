// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetOrder 取单个订单详情(管理端)。
//
// 与用户端的差别:多 user_id / username,且**不做归属校验** ——
// 管理端靠权限码授权(下一步做)。
//
// username 直接取快照列(order.username)—— 下单时已固化,不需要任何
// 额外 RPC,也不随用户改名而变。
func (l *GetOrderLogic) GetOrder(req *types.OrderIdReq) (*types.GetOrderResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.GetOrder(ctx, &v1_tradev1.GetOrderReq{
		OrderId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 → 400
		return nil, err
	}

	o := resp.GetOrder()
	return &types.GetOrderResp{
		OrderId:         o.GetOrderId(),
		OrderNo:         o.GetOrderNo(),
		OrderStatus:     o.GetOrderStatus(),
		TotalAmount:     o.GetTotalAmount(),
		PayAmount:       o.GetPayAmount(),
		PayMethod:       o.GetPayMethod(),
		PayTime:         formatTimestamp(o.GetPayTime()),
		AddressSnapshot: parseSnapshot(o.GetAddressSnapshot()),
		BuyerRemark:     o.GetBuyerRemark(),
		DetailList:      toOrderDetails(resp.GetDetails()),
		LogList:         toOrderLogs(resp.GetLogs()),
		CreatedAt:       formatTimestamp(o.GetCreatedAt()),
		Username:        o.GetUsername(),
		UserId:          o.GetUserId(),
	}, nil
}
