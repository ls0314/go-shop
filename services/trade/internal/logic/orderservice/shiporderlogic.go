package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ShipOrderLogic 管理端发货的协议适配层。
type ShipOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ShipOrderLogic) ShipOrder(in *v1_tradev1.ShipOrderReq) (*v1_tradev1.ShipOrderResp, error) {
	result, err := l.svcCtx.OrderService.ShipOrder(converter.FromProtoShipOrder(in))
	if err != nil {
		return &v1_tradev1.ShipOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ShipOrderResp{
		Order:          converter.ToProtoOrder(result.Order),
		ExpressCompany: result.ExpressCompany,
		TrackingNo:     result.TrackingNo,
	}, nil
}
