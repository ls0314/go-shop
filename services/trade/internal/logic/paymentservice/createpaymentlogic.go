package paymentservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreatePaymentLogic 发起支付的协议适配层。
type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePaymentLogic) CreatePayment(in *v1_tradev1.CreatePaymentReq) (*v1_tradev1.CreatePaymentResp, error) {
	record, err := l.svcCtx.PaymentService.CreatePayment(l.ctx,
		in.GetOrderId(), in.GetUserId(), in.GetPayMethod())
	if err != nil {
		return &v1_tradev1.CreatePaymentResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.CreatePaymentResp{
		PaymentId: record.PaymentId,
		PayNo:     record.PayNo,
		PayAmount: record.PayAmount,
		PayStatus: record.PayStatus,
		// PayUrl / QrCode 由**渠道**返回,本服务不构造 ——
		// 契约里为它们留了字段,接入真实渠道时在这里填
	}, nil
}
