package model

import "time"

// UserCartItem 购物车行结构体, 对应数据表user_cart_item。
type UserCartItem struct {
	CartItemId int64     `gorm:"primaryKey;column:cart_item_id" json:"cart_item_id"`
	UserId     int64     `gorm:"column:user_id" json:"user_id"`
	SkuId      int64     `gorm:"column:sku_id" json:"sku_id"`
	Quantity   int64     `gorm:"column:quantity" json:"quantity"`
	IsSelected bool      `gorm:"column:is_selected" json:"is_selected"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserCartItem) TableName() string {
	return "user_cart_item"
}

// CartItemView 购物车行 + 商品侧回填字段。
type CartItemView struct {
	UserCartItem
	SpuId      int64  `gorm:"column:spu_id" json:"spu_id"`
	SpuName    string `gorm:"column:spu_name" json:"spu_name"`
	SkuName    string `gorm:"column:sku_name" json:"sku_name"`
	SpecValues string `gorm:"column:spec_values" json:"spec_values"`
	MainImage  string `gorm:"column:main_image" json:"main_image"`
	// SkuImage SKU 自己的图,可能为空 —— 列表页优先用 MainImage(SPU 主图)
	SkuImage string  `gorm:"column:sku_image" json:"sku_image"`
	Price    float64 `gorm:"column:price" json:"price"`
	Stock    int64   `gorm:"column:stock" json:"stock"`
	// Available SKU/SPU 是否仍可购买。不可购买的行**仍要展示**(前端置灰),
	// 不静默丢弃 —— 否则用户会以为商品从购物车消失了
	Available bool `gorm:"-" json:"available"`
	// UnavailableReason 不可购买的原因文案(可能带库存数字,如"仅剩2件"),
	// 前端直接展示 —— 故不做成枚举。Available 为 true 时为空串
	UnavailableReason string `gorm:"-" json:"unavailable_reason"`
}
