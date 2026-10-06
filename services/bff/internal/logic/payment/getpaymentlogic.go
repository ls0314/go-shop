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

type GetPaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetPayment 按 pay_no 查支付记录。user_id 从 JWT 取,归属校验在服务端。
//
// 响应是内层 Payment 消息摊平(单体 response.GetPaymentResp 就是那样)。
func (l *GetPaymentLogic) GetPayment(req *types.GetPaymentReq) (*types.GetPaymentResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.PaymentRPC.GetPayment(ctx, &v1_tradev1.GetPaymentReq{
		UserId: userId,
		PayNo:  req.PayNo,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 支付记录不存在 / 不属于当前用户 → 400
		return nil, err
	}

	p := resp.GetPayment()
	return &types.GetPaymentResp{
		PaymentId: p.GetPaymentId(),
		PayNo:     p.GetPayNo(),
		OrderId:   p.GetOrderId(),
		OrderNo:   p.GetOrderNo(),
		PayMethod: p.GetPayMethod(),
		PayAmount: p.GetPayAmount(),
		PayStatus: p.GetPayStatus(),
		PayTime:   formatTimestamp(p.GetPayTime()),
		TradeNo:   p.GetTradeNo(),
	}, nil
}
