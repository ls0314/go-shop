package model

import "time"

// CartQuantityMax 单个 SKU 的加购数量上限。
const CartQuantityMax = 999

// CartQuantityMin 单个 SKU 的加购数量下限。
const CartQuantityMin = 1

// CartItemMaxCount 购物车行数上限(单 SKU 一行,故行数即"加购了几种商品")。
const CartItemMaxCount = 100

// ProductWithdraw 商品不可购买时的原因文案。
const ProductWithdraw = "商品已经下架"

// OrderPayTTL 支付时限的**默认值**。
const OrderPayTTL = 15 * time.Minute

// OrderScanBatch 超时扫描单轮处理上限。
const OrderScanBatch = 200

// 订单状态,与 ck_order_status 约束一致
const (
	OrderPendingPay = "pending_pay"
	OrderPaid       = "paid"
	OrderShipped    = "shipped"
	OrderCompleted  = "completed"
	OrderCancelled  = "cancelled"
)

// 订单操作类型,写 user_order_log.action
const (
	OrderActionCreate  = "create"
	OrderActionCancel  = "cancel"
	OrderActionPay     = "pay"
	OrderActionShip    = "ship"
	OrderActionConfirm = "confirm"
)

// 发件箱状态,与 status 列的取值一致
const (
	OutboxPending = "pending"
	OutboxSent    = "sent"
)

// 聚合根类型,写 aggregate_type
const OutboxAggregateOrder = "order"

// 事件类型,写 event_type
const OutboxEventOrderDelayCancel = "OrderDelayCancel"

// OutboxMaxRetry 投递重试上限。超过后不再重试(留在 pending 等人工/对账处理),
// 否则一条永远发不出去的消息会无限占用投递器的时间片。
const OutboxMaxRetry = 10

// 支付状态,与 ck_pay_status 约束一致
const (
	PayPending = "pending"
	PaySuccess = "success"
	PayFailed  = "failed"
	PayClosed  = "closed"
)

// 支付方式,与 ck_pay_method 约束一致
const (
	PayMethodMock   = "mock"
	PayMethodAlipay = "alipay"
	PayMethodWechat = "wechat"
)
