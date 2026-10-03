package model

// 库存变更类型,与 sys_product_stock_log.change_type 的 CHECK 约束一致
const (
	StockOrderLock     = "order_lock"
	StockPayDeduct     = "pay_deduct"
	StockOrderRelease  = "order_release"
	StockRefundRelease = "refund_release"
	// StockManualAdjust 管理端手动调整库存
	StockManualAdjust = "manual_adjust"
)

// SKU 状态
const (
	SkuStatusActive = "active"
)

// SPU 状态。
// 单体中 draft/withdrawn 是散落的字面量,拆分时统一提为常量;
// 取值与 sys_product_spu.spu_status 的 CHECK 约束一致。
const (
	// SpuStatusDraft 草稿
	SpuStatusDraft = "draft"
	// SpuStatusPublished 已上架
	SpuStatusPublished = "published"
	// SpuStatusWithdrawn 已下架
	SpuStatusWithdrawn = "withdrawn"
)

// 类目状态
const (
	// CategoryStatusActive 启用
	CategoryStatusActive = "active"
)
