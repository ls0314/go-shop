package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConfirmOrderLogic 确认收货的协议适配层
type ConfirmOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmOrderLogic {
	return &ConfirmOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmOrderLogic) ConfirmOrder(in *v1_tradev1.ConfirmOrderReq) (*v1_tradev1.ConfirmOrderResp, error) {
	order, err := l.svcCtx.OrderService.ConfirmOrder(in.GetOrderId(), in.GetUserId(), in.GetUsername())
	if err != nil {
		return &v1_tradev1.ConfirmOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ConfirmOrderResp{Order: converter.ToProtoOrder(order)}, nil
}
