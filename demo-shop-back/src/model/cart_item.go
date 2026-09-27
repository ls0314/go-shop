package model

import "time"

type UserCartItem struct {
	CartItemId int64     `gorm:"primary_key;autoIncrement;column:cart_item_id" json:"cart_item_id"`
	UserId     int64     `gorm:"column:user_id" json:"user_id"`
	SkuId      int64     `gorm:"column:sku_id" json:"sku_id"`
	Quantity   int64     `gorm:"column:quantity" json:"quantity"`
	IsSelected bool      `gorm:"column:is_selected" json:"is_selected"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserCartItem) TableName() string { return "user_cart_item" }
