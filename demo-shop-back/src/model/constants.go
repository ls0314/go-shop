package model

import "time"

const (
	StatusIdNotExist          = "ID不存在"
	StatusInternalServerError = "服务器错误"
	StatusBadRequest          = "请求参数错误"
	StatusNotExistRequest     = "请求内容不存在"
)

const (
	CartItemMinQuantity = 1
	CartItemMaxQuantity = 999
	CartMaxItemCount    = 100

	SkuStatusActive    = "active"
	SpuStatusPublished = "published"

	ProductWithdraw = "商品已经下架"

	OrderPayTTL = 15 * time.Minute
)

const (
	StockManualAdjust  = "manual_adjust"
	StockOrderLock     = "order_lock"
	StockPayDeduct     = "pay_deduct"
	StockOrderRelease  = "order_release"
	StockRefundRelease = "refund_release"
)

const (
	OrderPendingPay = "pending_pay"
	OrderPaid       = "paid"
	OrderShipped    = "shipped"
	OrderCompleted  = "completed"
	OrderCancelled  = "cancelled"

	OrderCreate     = "create"
	OrderPay        = "pay"
	OrderCancel     = "cancel"
	OrderShip       = "ship"
	OrderConfirm    = "confirm"
	OrderAutoCancel = "auto_cancel"
)

// 支付状态
const (
	PayPending = "pending"
	PaySuccess = "success"
	PayFailed  = "failed"
	PayClosed  = "closed"
)

// 支付方式
const (
	PayMethodMock   = "mock"
	PayMethodWechat = "wechat"
	PayMethodAlipay = "alipay"
)
const (
	CouponUnused = "unused"
)

// 操作日志模块名（OperationLogMiddleware 挂载参数）
const (
	LogModuleProduct    = "商品管理"
	LogModuleOrder      = "订单管理"
	LogModuleInventory  = "库存管理"
	LogModuleCategory   = "类目管理"
	LogModulePermission = "权限管理"
	LogModuleUser       = "用户管理"
)

// outbox状态
const (
	OutboxPending = "pending"
	OutboxSent    = "sent"
)

// outbox 事件类型
const (
	OutboxAggregateOrder        = "order"            // 聚合根:订单
	OutboxEventOrderDelayCancel = "OrderDelayCancel" // 订单延迟取消意图
)
