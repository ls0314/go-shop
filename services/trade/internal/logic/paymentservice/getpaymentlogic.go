package paymentservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetPaymentLogic 查支付流水的协议适配层
type GetPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaymentLogic) GetPayment(in *v1_tradev1.GetPaymentReq) (*v1_tradev1.GetPaymentResp, error) {
	record, err := l.svcCtx.PaymentService.GetPayment(in.GetUserId(), in.GetPayNo())
	if err != nil {
		return &v1_tradev1.GetPaymentResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetPaymentResp{Payment: converter.ToProtoPayment(record)}, nil
}
