package pay

import (
	"context"
	"demo-shop-back/src/model/requset"
)

// ============================================================
// 支付网关 — 统一接口及数据模型
// ============================================================
//
// 设计目的：通过策略模式将支付方式的差异封装在网关实现层，
// Service 层只依赖 PayGateway 接口，新增支付渠道只需实现接口并注册。

// ============================================================
// 请求/响应数据结构
// ============================================================

// PayRequest 创建支付的统一请求参数
// 各网关实现根据自己的 API 要求将本结构体转为对应的参数格式
type PayRequest struct {
	PayNo    string // 支付流水号（内部生成，全局唯一）
	OrderNo  string // 订单号（用于展示给用户）
	Amount   string // 支付金额（字符串格式，避免浮点精度问题，例："6999.00"）
	Subject  string // 商品/订单描述（微信：body，支付宝：subject）
	ExpireAt string // 支付过期时间（RFC3339 字符串）
}

// PayResponse 创建支付的统一响应
// mock 模式各字段为空；微信/支付宝返回支付链接或二维码供前端展示
type PayResponse struct {
	PayUrl string // 支付页面链接（支付宝返回，前端跳转）
	QrCode string // 二维码数据（微信扫码支付返回，前端生成二维码）
	PayNo  string // 支付流水号（原样回传，便于日志追踪）
}

// CallbackData 支付回调解析后的统一数据结构
// 各网关的 ParsePayment 方法将第三方原始通知转为此结构体
type CallbackData struct {
	PayNo     string  // 支付流水号（从回调中提取）
	TradeNo   string  // 第三方支付交易号（微信的 transaction_id / 支付宝的 trade_no）
	Amount    float64 // 回调中的支付金额（用于校验，mock 模式不校验）
	Status    string  // 支付状态（success/failed）
	PayMethod string  // 支付方式标识（回传给 Service 层，便于日志和后续处理）
	RawData   string  // 回调原始数据（存入 notify_log，用于排障/对账）
}

// ============================================================
// PayGateway 统一支付网关接口
// ============================================================
//
// 实现类：MockPayGateway / WechatPayGateway / AlipayGateway
// 注册入口：registry.go 的 InitGateways()
// 调用方：payment_service.go 的 CreatePayment / HandleCallback
type PayGateway interface {

	// Name 返回支付方式标识
	// 对应 model.PayMethodMock / PayMethodWechat / PayMethodAlipay
	Name() string

	// CreatePayment 创建支付订单，返回前端所需的支付参数
	// mock：直接返回空 PayResponse（前端调回调接口完成支付）
	// 微信：调用统一下单 API，返回 code_url（前端生成扫码二维码）
	// 支付宝：调用 pageExecute，返回支付页面 URL（前端跳转）
	CreatePayment(ctx context.Context, req PayRequest) (*PayResponse, error)

	// ParsePayment 解析并验证支付平台回调通知
	// mock：直接透传参数，不做签名验证
	// 微信：解析 XML → 验证 HMAC-SHA256/RSA 签名 → 返回 CallbackData
	// 支付宝：解析 form-data → RSA2 验签 → 调用 notify_id 验证接口（防重放）→ 返回 CallbackData
	ParsePayment(ctx context.Context, rawParams requset.CallbackPaymentReq) (*CallbackData, error)

	// QueryPayment 主动查询支付结果（补偿/对账用）
	// 用于处理回调丢失的情况，或前台轮询支付状态
	// mock：直接返回 success（无需查询）
	// 微信：调 orderquery 接口
	// 支付宝：调 alipay.trade.query 接口
	QueryPayment(ctx context.Context, payNo string) (*CallbackData, error)
}
