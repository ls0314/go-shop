package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateOrderLogic 下单的协议适配层。
type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *v1_tradev1.CreateOrderReq) (*v1_tradev1.CreateOrderResp, error) {
	result, err := l.svcCtx.OrderService.CreateOrder(l.ctx, converter.FromProtoCreateOrder(in))
	if err != nil {
		return &v1_tradev1.CreateOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return converter.ToProtoCreateOrderResp(result), nil
}
