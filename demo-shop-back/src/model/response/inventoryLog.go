package response

import "demo-shop-back/src/model"

type InventoryLogList struct {
	model.SysProductStockLog
	SkuName string `json:"sku_name"`
	SpuName string `json:"spu_name"`
}
type InventoryLogResp struct {
	List     []InventoryLogList `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type InventoryAdjustResp struct {
	BeforeStock int64 `json:"before_stock"`
	AfterStock  int64 `json:"after_stock"`
}

type InventoryWarnResp struct {
	SkuId     int64  `gorm:"column:sku_id" json:"sku_id"`
	SpuName   string `gorm:"column:spu_name" json:"spu_name"`
	SkuName   string `gorm:"column:sku_name" json:"sku_name"`
	Stock     int64  `gorm:"column:stock" json:"stock"`
	LockStock int64  `gorm:"column:lock_stock" json:"lock_stock"`
	SoldCount int64  `gorm:"column:sold_count" json:"sold_count"`
	SkuStatus string `gorm:"column:sku_status" json:"sku_status"`
}
