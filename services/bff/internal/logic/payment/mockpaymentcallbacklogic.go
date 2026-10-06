// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MockPaymentCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockPaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockPaymentCallbackLogic {
	return &MockPaymentCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MockPaymentCallback 模拟渠道支付回调(**免鉴权**路由)。
//
// pay_no 必填;trade_no 可选。
//
// 错误语义:
//
//	ResultInfra → 503。渠道会重试,这是我们想要的(依赖抖动)
//	ResultBiz   → 400。流水不存在 / 金额不符 —— 重试无用,不该让渠道
//	              无限重投
//
// accepted=false 时按业务失败回 400,与单体一致(utils.Error(c, 500, errMsg))。
// **状态码与单体不同**(单体一律 500),但 400 更准确:它表达"这次回调
// 的内容不成立",而 500 会让渠道把它当成我方故障并持续重投。
//
// idempotent=true 表示重复回调且未产生新变更 —— 这是正常结果(渠道
// 重投是常态),照常回 200,只是记一条日志便于观察重投频率。
func (l *MockPaymentCallbackLogic) MockPaymentCallback(req *types.MockCallbackReq) (*types.Empty, error) {
	if req.PayNo == "" {
		return nil, errInvalidParam
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.PaymentRPC.HandlePaymentCallback(ctx, &v1_tradev1.HandlePaymentCallbackReq{
		PayNo:   req.PayNo,
		TradeNo: req.TradeNo,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		// 504/503 → 渠道会重试(见上方说明)
		return nil, err
	case rpc.ResultBiz:
		// accepted=false 或 error_msg 非空:校验未通过,重试无用
		return nil, err
	}

	if !resp.GetAccepted() {
		return nil, errInvalidParam
	}

	if resp.GetIdempotent() {
		logx.WithContext(l.ctx).Infof("支付回调重复(未产生变更): pay_no=%s", req.PayNo)
	}

	return &types.Empty{}, nil
}
