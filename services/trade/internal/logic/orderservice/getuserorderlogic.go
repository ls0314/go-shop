package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetUserOrderLogic 用户端订单详情的协议适配层
type GetUserOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOrderLogic {
	return &GetUserOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserOrderLogic) GetUserOrder(in *v1_tradev1.GetUserOrderReq) (*v1_tradev1.GetUserOrderResp, error) {
	view, err := l.svcCtx.OrderService.GetUserOrder(in.GetOrderId(), in.GetUserId())
	if err != nil {
		return &v1_tradev1.GetUserOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetUserOrderResp{
		Order:   converter.ToProtoOrder(view.Order),
		Details: converter.ToProtoOrderDetails(view.Details),
		Logs:    converter.ToProtoOrderLogs(view.Logs),
	}, nil
}
