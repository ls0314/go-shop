package response

import "gorm.io/datatypes"

type CartItemListResp struct {
	CartItemId        int64             `gorm:"column:cart_item_id" json:"cart_item_id"`
	SkuId             int64             `gorm:"column:sku_id" json:"sku_id"`
	SpuId             int64             `gorm:"column:spu_id" json:"spu_id"`
	SpuName           string            `gorm:"column:spu_name" json:"spu_name"`
	MainImage         string            `gorm:"column:main_image" json:"main_image"`
	SkuName           string            `gorm:"column:sku_name" json:"sku_name"`
	SpecValues        datatypes.JSONMap `gorm:"column:spec_values" json:"spec_values"`
	SkuImage          string            `gorm:"column:sku_image" json:"sku_image"`
	Price             float64           `gorm:"column:price" json:"price"`
	Stock             int64             `gorm:"column:stock" json:"stock"`
	Quantity          int64             `gorm:"column:quantity" json:"quantity"`
	IsSelected        bool              `gorm:"column:is_selected" json:"is_selected"`
	Subtotal          float64           `gorm:"-" json:"subtotal"`
	IsAvailable       bool              `gorm:"-" json:"is_available"`
	UnavailableReason string            `gorm:"-" json:"unavailable_reason"`
	UserId            int64             `gorm:"column:user_id" json:"-"`
	SkuStatus         string            `gorm:"column:sku_status" json:"-"`
	SpuStatus         string            `gorm:"column:spu_status" json:"-"`
}

type CartItemCreateResp struct {
	CartItemId int64 `json:"cart_item_id"`
	Quantity   int64 `json:"quantity"`
}

type CartItemPayResp struct {
	Items            []CartItemListResp `json:"items"`
	TotalCount       int64              `json:"total_count"`
	TotalQuantity    int64              `json:"total_quantity"`
	TotalAmount      float64            `json:"total_amount"`
	HasUnavailable   bool               `json:"has_unavailable"`
	UnavailableItems []CartItemListResp `json:"unavailable_items"`
}
