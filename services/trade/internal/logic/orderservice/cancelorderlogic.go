package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CancelOrderLogic 取消订单的协议适配层。
type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelOrderLogic) CancelOrder(in *v1_tradev1.CancelOrderReq) (*v1_tradev1.CancelOrderResp, error) {
	result, err := l.svcCtx.OrderService.CancelOrder(l.ctx, converter.FromProtoCancelOrder(in))
	if err != nil {
		return &v1_tradev1.CancelOrderResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.CancelOrderResp{
		Order: converter.ToProtoOrder(result.Order),
		// Compensated=false 表示"订单已取消,但库存释放/退券还没成功" ——
		// 上游据此告警,而不是当成失败
		Compensated: result.Compensated,
	}, nil
}
