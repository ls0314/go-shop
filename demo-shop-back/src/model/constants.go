package model

const (
	CartItemMinQuantity = 1
	CartItemMaxQuantity = 999
	CartMaxItemCount    = 100

	SkuStatusActive    = "active"
	SpuStatusPublished = "published"

	ProductWithdraw = "商品已经下架"
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
