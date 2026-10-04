package paymentservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// HandlePaymentCallbackLogic 渠道回调的协议适配层。
type HandlePaymentCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandlePaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlePaymentCallbackLogic {
	return &HandlePaymentCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandlePaymentCallbackLogic) HandlePaymentCallback(in *v1_tradev1.HandlePaymentCallbackReq) (*v1_tradev1.HandlePaymentCallbackResp, error) {
	result, err := l.svcCtx.PaymentService.HandleCallback(l.ctx, in.GetPayNo(), in.GetTradeNo())
	if err != nil {
		// 业务失败(流水不存在/金额不符)与向前补偿信号都走这里:
		// 前者 err 会被 bizErrMsg 翻成文案,后者文案为空但 err 非 nil,
		// 渠道据 err 决定要不要重试
		return &v1_tradev1.HandlePaymentCallbackResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.HandlePaymentCallbackResp{
		Accepted:   result.Accepted,
		Idempotent: result.Idempotent,
	}, nil
}
