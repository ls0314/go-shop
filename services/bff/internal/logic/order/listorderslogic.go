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

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListOrders 分页查询订单(管理端)。
//
// 列表项比用户端多 username / receiver_name / receiver_phone,三者都
// **不需要额外 RPC**:
//
//	username        取订单表的快照列(order.username,下单时固化)
//	receiver_*      从 address_snapshot 那个 JSON 里解析
//
// 时间筛选(start_time / end_time)由 BFF 解析成 Timestamp。
// 响应的 page / page_size 回写兜底后的值。
func (l *ListOrdersLogic) ListOrders(req *types.ListOrdersReq) (*types.ListOrdersResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	startTime, err := parseTimeParam(req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := parseTimeParam(req.EndTime)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.ListOrders(ctx, &v1_tradev1.ListOrdersReq{
		Page:        int32(page),
		PageSize:    int32(pageSize),
		OrderStatus: req.OrderStatus,
		OrderNo:     req.OrderNo,
		StartTime:   startTime,
		EndTime:     endTime,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	items := resp.GetItems()
	list := make([]types.AdminOrderListItem, 0, len(items))
	for _, o := range items {
		// 收货人从地址快照里取
		snap, _ := parseSnapshot(o.GetAddressSnapshot()).(map[string]interface{})
		receiver, _ := snap["receiver_name"].(string)
		phone, _ := snap["receiver_phone"].(string)

		list = append(list, types.AdminOrderListItem{
			OrderId:     o.GetOrderId(),
			OrderNo:     o.GetOrderNo(),
			OrderStatus: o.GetOrderStatus(),
			TotalAmount: o.GetTotalAmount(),
			PayAmount:   o.GetPayAmount(),
			PayMethod:   o.GetPayMethod(),
			// 用户名快照,下单时已固化 —— 不随用户改名而变
			UserName:      o.GetUsername(),
			ReceiverName:  receiver,
			ReceiverPhone: phone,
			CreatedAt:     formatTimestamp(o.GetCreatedAt()),
		})
	}

	return &types.ListOrdersResp{
		Page:     page,
		PageSize: pageSize,
		Total:    resp.GetTotal(),
		List:     list,
	}, nil
}
