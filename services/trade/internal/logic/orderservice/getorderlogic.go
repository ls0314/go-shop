package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetOrderLogic 管理端订单详情的协议适配层。
type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *v1_tradev1.GetOrderReq) (*v1_tradev1.GetOrderResp, error) {
	view, err := l.svcCtx.OrderService.GetOrder(in.GetOrderId())
	if err != nil {
		return &v1_tradev1.GetOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetOrderResp{
		Order:   converter.ToProtoOrder(view.Order),
		Details: converter.ToProtoOrderDetails(view.Details),
		Logs:    converter.ToProtoOrderLogs(view.Logs),
		// Username 由调用方(user 域)回填:订单表只存 user_id,
		// 跨库取不到用户名 —— 契约里为它留了字段,本服务不假装能填
	}, nil
}
