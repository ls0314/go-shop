// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreatePayment 发起支付。:id 是 order_id,user_id 从 JWT 取。
//
// 响应对应单体 response.CreatePaymentResp。注意 pay_url / qr_code 在
// 单体带 `omitempty`,而 goctl 不支持 —— 这里空时也出现(值为空串)。
// 前端若用 `'pay_url' in data` 判断会看到行为差异,用 truthy 判断则无影响。
func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (*types.CreatePaymentResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.PaymentRPC.CreatePayment(ctx, &v1_tradev1.CreatePaymentReq{
		OrderId:   req.Id,
		UserId:    userId,
		PayMethod: req.PayMethod,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 / 不属于当前用户 / 状态不可支付 / 金额异常 → 400
		return nil, err
	}

	return &types.CreatePaymentResp{
		PaymentId: resp.GetPaymentId(),
		PayNo:     resp.GetPayNo(),
		PayAmount: resp.GetPayAmount(),
		PayStatus: resp.GetPayStatus(),
		PayUrl:    resp.GetPayUrl(),
		QrCode:    resp.GetQrCode(),
	}, nil
}
