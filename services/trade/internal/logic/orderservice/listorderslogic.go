package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListOrdersLogic 管理端订单列表的协议适配层
type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *v1_tradev1.ListOrdersReq) (*v1_tradev1.ListOrdersResp, error) {
	page, err := l.svcCtx.OrderService.ListOrders(converter.FromProtoListOrders(in))
	if err != nil {
		return &v1_tradev1.ListOrdersResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ListOrdersResp{
		Items:    converter.ToProtoOrderList(page.Items),
		Total:    page.Total,
		Page:     int32(page.Page),
		PageSize: int32(page.PageSize),
	}, nil
}
