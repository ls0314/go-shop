package pay

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"errors"
)

// ============================================================
// 模拟支付网关 — PayGateway 接口的 mock 实现
// ============================================================
//
// 使用场景：开发/测试阶段，不对接真实支付平台。
// 特点：不生成支付页面、不做签名验证、前端主动调用回调接口完成支付闭环。
//
// 注册时机：InitGateways() 中始终注册，无需额外配置。

// MockPayGateway 模拟支付网关
// 空结构体 — mock 模式无需存储商户号、密钥等配置
type MockPayGateway struct{}

// NewMockPayGateway 创建模拟支付网关实例
func NewMockPayGateway() *MockPayGateway {
	return &MockPayGateway{}
}

// Name 返回支付方式标识
// 返回值：model.PayMethodMock（"mock"）
func (m *MockPayGateway) Name() string {
	return model.PayMethodMock
}

// CreatePayment mock 模式不调用第三方 API，直接返回空 PayResponse
// 前端拿到响应后，由用户主动 POST /api/v1/pay/callback/mock 完成支付
func (m *MockPayGateway) CreatePayment(ctx context.Context, req PayRequest) (*PayResponse, error) {
	return &PayResponse{
		PayNo: req.PayNo,
	}, nil
}

// ParsePayment 解析模拟支付回调参数
// mock 模式不做签名验证，直接校验 pay_no 非空后透传参数
// 请求参数（Body JSON）：
//
//	pay_no   - 支付流水号，string，必填
//	trade_no  - 模拟交易号，string，可选
//
// 返回值：CallbackData，状态固定为 success
func (m *MockPayGateway) ParsePayment(ctx context.Context, rawParams requset.CallbackPaymentReq) (*CallbackData, error) {
	payNo := rawParams.PayNo
	if payNo == "" {
		return nil, errors.New("pay_no 不能为空")
	}
	return &CallbackData{
		PayNo:     payNo,
		TradeNo:   rawParams.TradeNo,
		Status:    model.PaySuccess,
		PayMethod: model.PayMethodMock,
	}, nil
}

// QueryPayment mock 模式不查询第三方，直接返回 success
// 用于补偿/对账场景，或前台轮询支付状态的兜底
func (m *MockPayGateway) QueryPayment(ctx context.Context, payNo string) (*CallbackData, error) {
	return &CallbackData{
		PayNo:     payNo,
		Status:    model.PaySuccess,
		PayMethod: model.PayMethodMock,
	}, nil
}
