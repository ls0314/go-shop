package response

import "time"

type CreatePaymentResp struct {
	PaymentId int64   `json:"payment_id"`
	PayNo     string  `json:"pay_no"`
	PayAmount float64 `json:"pay_amount"`
	PayStatus string  `json:"pay_status"`
	PayUrl    string  `json:"pay_url,omitempty"` // 支付链接（微信/支付宝返回）
	QrCode    string  `json:"qr_code,omitempty"` // 二维码数据
}

// CallbackPaymentResp 支付回调响应
type CallbackPaymentResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GetPaymentResp struct {
	PaymentId int64     `gorm:"column:payment_id" json:"payment_id"`
	PayNo     string    `gorm:"column:pay_no" json:"pay_no"`
	OrderId   int64     `gorm:"column:order_id" json:"order_id"`
	OrderNo   string    `gorm:"column:order_no" json:"order_no"`
	PayMethod string    `gorm:"column:pay_method" json:"pay_method"`
	PayAmount float64   `gorm:"column:pay_amount" json:"pay_amount"`
	PayStatus string    `gorm:"column:pay_status" json:"pay_status"`
	PayTime   time.Time `gorm:"column:pay_time" json:"pay_time"`
	TradeNo   string    `gorm:"column:trade_no" json:"trade_no"`
}

type GetPaymentList struct {
	PaymentId int64     `json:"payment_id"`
	PayNo     string    `json:"pay_no"`
	OrderNo   string    `json:"order_no"`
	Username  string    `json:"username"`
	PayMethod string    `json:"pay_method"`
	PayAmount float64   `json:"pay_amount"`
	PayStatus string    `json:"pay_status"`
	PayTime   time.Time `json:"pay_time"`
}

type GetPaymentListResp struct {
	List     []GetPaymentList `json:"list"`
	Total    int64            `json:"total"`
	PageSize int              `json:"page_size"`
	Page     int              `json:"page"`
}
