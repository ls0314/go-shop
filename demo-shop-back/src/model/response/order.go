package response

import (
	"time"

	"gorm.io/datatypes"
)

type CreateOrderResp struct {
	OrderId     int64     `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	PayAmount   float64   `json:"pay_amount"`
	TotalAmount float64   `json:"total_amount"`
	OrderStatus string    `json:"order_status"`
	PayExpireAt time.Time `json:"pay_expire_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserGetOrderList struct {
	OrderId     int64     `gorm:"order_id" json:"order_id"`
	OrderNo     string    `gorm:"order_no" json:"order_no"`
	OrderStatus string    `gorm:"order_status" json:"order_status"`
	TotalAmount float64   `gorm:"total_amount" json:"total_amount"`
	PayAmount   float64   `gorm:"pay_amount" json:"pay_amount"`
	DetailCount int64     `gorm:"detail_count" json:"detail_count"`
	FirstImage  string    `gorm:"first_image" json:"first_image"`
	CreatedAt   time.Time `gorm:"created_at" json:"created_at"`
}

type UserGetOrderListResp struct {
	List     []UserGetOrderList `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type UserGetOrderResp struct {
	OrderId         int64                `json:"order_id"`
	OrderNo         string               `json:"order_no"`
	OrderStatus     string               `json:"order_status"`
	TotalAmount     float64              `json:"total_amount"`
	PayAmount       float64              `json:"pay_amount"`
	PayMethod       string               `json:"pay_method"`
	PayTime         time.Time            `json:"pay_time"`
	AddressSnapshot datatypes.JSON       `json:"address_snapshot"`
	BuyerRemark     string               `json:"buyer_remark"`
	DetailList      []UserGetOrderDetail `json:"detail_list"`
	LogList         []UserGetOrderLog    `json:"log_list"`
	CreatedAt       time.Time            `json:"created_at"`
}

type UserGetOrderDetail struct {
	DetailId   int64   `json:"detail_id"`
	SkuId      int64   `json:"sku_id"`
	SpuName    string  `json:"spu_name"`
	SkuName    string  `json:"sku_name"`
	SpecValues string  `json:"spec_values"`
	MainImage  string  `json:"main_image"`
	Quantity   int64   `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type UserGetOrderLog struct {
	LogId       int64     `json:"log_id"`
	OrderId     int64     `json:"order_id"`
	OrderStatus string    `json:"order_status"`
	Action      string    `json:"action"`
	Operator    string    `json:"operator"`
	Detail      string    `json:"detail"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetOrderListResp struct {
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Total    int64          `json:"total"`
	List     []GetOrderList `json:"list"`
}
type GetOrderList struct {
	OrderId       int64     `json:"order_id"`
	OrderNo       string    `json:"order_no"`
	OrderStatus   string    `json:"order_status"`
	TotalAmount   float64   `json:"total_amount"`
	PayAmount     float64   `json:"pay_amount"`
	PayMethod     string    `json:"pay_method"`
	UserName      string    `json:"user_name"`
	ReceiverName  string    `json:"receiver_name"`
	ReceiverPhone string    `json:"receiver_phone"`
	CreatedAt     time.Time `json:"created_at"`
}

type GetOrderResp struct {
	OrderId         int64                `gorm:"order_id" json:"order_id"`
	OrderNo         string               `gorm:"order_no" json:"order_no"`
	OrderStatus     string               `gorm:"order_status" json:"order_status"`
	TotalAmount     float64              `gorm:"total_amount" json:"total_amount"`
	PayAmount       float64              `gorm:"pay_amount" json:"pay_amount"`
	PayMethod       string               `gorm:"pay_method" json:"pay_method"`
	PayTime         time.Time            `gorm:"pay_time" json:"pay_time"`
	AddressSnapshot datatypes.JSON       `gorm:"address_snapshot" json:"address_snapshot"`
	BuyerRemark     string               `gorm:"buyer_remark" json:"buyer_remark"`
	DetailList      []UserGetOrderDetail `gorm:"-" json:"detail_list"`
	LogList         []UserGetOrderLog    `gorm:"-" json:"log_list"`
	CreatedAt       time.Time            `gorm:"created_at" json:"created_at"`
	Username        string               `gorm:"username" json:"username"`
	UserId          int64                `gorm:"user_id" json:"user_id"`
}

type OrderShipResp struct {
	OrderId        int64  `json:"order_id"`
	OrderNo        string `json:"order_no"`
	OrderStatus    string `json:"order_status"`
	ExpressCompany string `json:"express_company"`
	TrackingNO     string `json:"tracking_no"`
}

type OrderStatusResp struct {
	OrderId     int64  `json:"order_id"`
	OrderNo     string `json:"order_no"`
	OrderStatus string `json:"order_status"`
}
