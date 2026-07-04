package requset

import (
	"demo-shop-back/src/model"

	"gorm.io/datatypes"
)

// SpuQueryReq 获取商品列表请求
type SpuQueryReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	SpuName    string `form:"spu_name"`
	CategoryId *int64 `form:"category_id"`
	SpuStatus  string `form:"spu_status"`
	Brand      string `form:"brand"`
	Sort       string `form:"sort"`
}

// FullUpdateProductReq 全量更新商品请求
// SKU 对比规则：传入列表中有 sku_id 的 → 更新，无 sku_id 的 → 新增，DB 有但列表无的 → 软删
type FullUpdateProductReq struct {
	SpuName        string                      `json:"spu_name"`
	CategoryId     *int64                      `json:"category_id"`
	Brand          string                      `json:"brand"`
	Description    string                      `json:"description"`
	MainImage      string                      `json:"main_image"`
	SpecTemplate   *datatypes.JSON             `json:"spec_template"`
	Priority       *int64                      `json:"priority"`
	SkuList        *[]model.SysProductSku      `json:"sku_list"`
	ImageList      *[]model.SysProductSpuImage `json:"image_list"`
	DeleteImageIds []int64                     `json:"delete_image_ids"`
}
