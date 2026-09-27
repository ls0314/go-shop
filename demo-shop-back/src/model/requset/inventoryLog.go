package requset

import "time"

type InventoryLogReq struct {
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
	SkuId      *int64     `form:"sku_id"`
	SpuId      *int64     `form:"spu_id"`
	ChangeType string     `form:"change_type"`
	StartTime  *time.Time `form:"start_time"`
	EndTime    *time.Time `form:"end_time"`
}

type InventoryAdjustReq struct {
	SkuId     int64  `json:"sku_id"`
	ChangeQty int64  `json:"change_qty"`
	Remark    string `json:"remark"`
}

type InventoryWarnReq struct {
	Threshold int    `form:"threshold"`
	SpuStatus string `form:"spu_status"`
}
