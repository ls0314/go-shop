package model

import (
	"time"

	"gorm.io/datatypes"
)

// SysProductSku 商品SKU
type SysProductSku struct {
	SkuId      int64             `gorm:"primaryKey;autoIncrement;column:sku_id" json:"sku_id"`
	SpuId      int64             `gorm:"column:spu_id" json:"spu_id"`
	SkuName    string            `gorm:"column:sku_name" json:"sku_name"`
	SpecValues datatypes.JSONMap `gorm:"column:spec_values" json:"spec_values"`
	Price      float64           `gorm:"column:price" json:"price"`
	CostPrice  float64           `gorm:"column:cost_price" json:"cost_price"`
	Stock      int64             `gorm:"column:stock" json:"stock"`
	LockStock  int64             `gorm:"column:lock_stock" json:"lock_stock"`
	SoldCount  int64             `gorm:"column:sold_count" json:"sold_count"`
	SkuCode    string            `gorm:"column:sku_code" json:"sku_code"`
	SkuImage   string            `gorm:"column:sku_image" json:"sku_image"`
	SkuStatus  string            `gorm:"column:sku_status" json:"sku_status"`
	IsDeleted  bool              `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt  time.Time         `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"column:updated_at" json:"updated_at"`
	CreateBy   int64             `gorm:"column:create_by" json:"create_by"`
	UpdateBy   int64             `gorm:"column:update_by" json:"update_by"`
}

func (SysProductSku) TableName() string { return "sys_product_sku" }

// SysProductStockLog 库存变更流水
type SysProductStockLog struct {
	LogId       int64  `gorm:"primaryKey;column:log_id" json:"log_id"`
	SkuId       int64  `gorm:"column:sku_id" json:"sku_id"`
	ChangeType  string `gorm:"column:change_type" json:"change_type"`
	ChangeQty   int64  `gorm:"column:change_qty" json:"change_qty"`
	BeforeStock int64  `gorm:"column:before_stock" json:"before_stock"`
	AfterStock  int64  `gorm:"column:after_stock" json:"after_stock"`
	BeforeLock  int64  `gorm:"column:before_lock" json:"before_lock"`
	AfterLock   int64  `gorm:"column:after_lock" json:"after_lock"`
	// IdempotencyKey 幂等键(调用方生成,全局唯一)。
	//
	// 四操作(order_lock / pay_deduct / order_release / refund_release)判重的**唯一依据**,
	// 靠部分唯一索引 uk_stock_log_idem (idempotency_key, sku_id, change_type) 实现。
	// 手工调整不带它(NULL 不参与唯一性判断,故可重复调整 —— 这是刻意的,
	// 手动调整没有"同一次操作"的概念)。
	IdempotencyKey string `gorm:"column:idempotency_key" json:"idempotency_key"`
	// OrderNo 关联订单号,**仅用于追溯**(运维按单号查这张单动了哪些库存),
	// 不参与任何幂等判断。允许为空:退款等场景调用方可能拿不到单号。
	OrderNo string `gorm:"column:order_no" json:"order_no"`
	// OrderId 关联订单ID。**只有历史流水有值** —— 拆库后 order_id 由 trade 侧
	// 建单时分配,而流水在锁库存时就写了(早于建单)。用 OrderNo 追溯
	OrderId   *int64    `gorm:"column:order_id" json:"order_id"`
	Remark    string    `gorm:"column:remark" json:"remark"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	CreateBy  int64     `gorm:"column:create_by" json:"create_by"`
}

func (SysProductStockLog) TableName() string { return "sys_product_stock_log" }
