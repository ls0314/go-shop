package model

import (
	"time"

	"gorm.io/datatypes"
)

type UserOrder struct {
	OrderId         int64          `gorm:"column:order_id;primary_key;AUTO_INCREMENT" json:"order_id"`
	OrderNo         string         `gorm:"column:order_no" json:"order_no"`
	UserId          int64          `gorm:"column:user_id" json:"user_id"`
	OrderStatus     string         `gorm:"column:order_status" json:"order_status"`
	TotalAmount     float64        `gorm:"column:total_amount" json:"total_amount"`
	PayAmount       float64        `gorm:"column:pay_amount" json:"pay_amount"`
	PayMethod       string         `gorm:"column:pay_method" json:"pay_method"`
	PayTime         time.Time      `gorm:"column:pay_time" json:"pay_time"`
	AddressSnapshot datatypes.JSON `gorm:"column:address_snapshot" json:"address_snapshot"`
	BuyerRemark     string         `gorm:"column:buyer_remark" json:"buyer_remark"`
	IdempotentKey   string         `gorm:"column:idempotent_key" json:"idempotent_key"`
	IsDeleted       bool           `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt       time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DetailCount     int64          `gorm:"column:detail_count" json:"detail_count"`
	FirstImage      string         `gorm:"column:first_image" json:"first_image"`
}

func (UserOrder) TableName() string { return "user_order_master" }

type UserOrderDetail struct {
	DetailId   int64             `gorm:"column:detail_id;primary_key;AUTO_INCREMENT" json:"detail_id"`
	OrderId    int64             `gorm:"column:order_id" json:"order_id"`
	SkuId      int64             `gorm:"column:sku_id" json:"sku_id"`
	SpuName    string            `gorm:"column:spu_name" json:"spu_name"`
	SkuName    string            `gorm:"column:sku_name" json:"sku_name"`
	SpecValues datatypes.JSONMap `gorm:"column:spec_values" json:"spec_values"`
	MainImage  string            `gorm:"column:main_image" json:"main_image"`
	Quantity   int64             `gorm:"column:quantity" json:"quantity"`
	UnitPrice  float64           `gorm:"column:unit_price" json:"unit_price"`
	TotalPrice float64           `gorm:"column:total_price" json:"total_price"`
	CreatedAt  time.Time         `gorm:"column:created_at" json:"created_at"`
}

func (UserOrderDetail) TableName() string { return "user_order_detail" }

type UserOrderLog struct {
	LogId       int64     `gorm:"column:log_id;primary_key;AUTO_INCREMENT" json:"log_id"`
	OrderId     int64     `gorm:"column:order_id" json:"order_id"`
	OrderStatus string    `gorm:"column:order_status" json:"order_status"`
	Action      string    `gorm:"column:action" json:"action"`
	Operator    string    `gorm:"column:operator" json:"operator"`
	Detail      string    `gorm:"column:detail" json:"detail"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (UserOrderLog) TableName() string { return "user_order_log" }

type AddressSnap struct {
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	PostalCode    string `json:"postal_code"`
}
