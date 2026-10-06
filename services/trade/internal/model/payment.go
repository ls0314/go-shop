package model

import (
	"time"

	"gorm.io/datatypes"
)

// UserPaymentRecord 支付流水结构体, 对应数据表user_payment_record
type UserPaymentRecord struct {
	PaymentId int64  `gorm:"primaryKey;column:payment_id" json:"payment_id"`
	PayNo     string `gorm:"column:pay_no" json:"pay_no"`
	OrderId   int64  `gorm:"column:order_id" json:"order_id"`
	UserId    int64  `gorm:"column:user_id" json:"user_id"`
	// Username 支付时的用户名**快照**(建支付时从订单读,与订单一致)。
	Username  string         `gorm:"column:username" json:"username"`
	PayMethod string         `gorm:"column:pay_method" json:"pay_method"`
	PayAmount float64        `gorm:"column:pay_amount" json:"pay_amount"`
	PayStatus string         `gorm:"column:pay_status" json:"pay_status"`
	TradeNo   string         `gorm:"column:trade_no" json:"trade_no"`
	PayTime   time.Time      `gorm:"column:pay_time" json:"pay_time"`
	NotifyLog datatypes.JSON `gorm:"column:notify_log;type:jsonb" json:"notify_log"`
	ExpireAt  time.Time      `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (UserPaymentRecord) TableName() string {
	return "user_payment_record"
}

// PaymentView 支付流水 + 联表取到的订单号。
//
// Username 不再走 JOIN —— 它是 user_payment_record 自己的快照列,
// 由内嵌的 UserPaymentRecord 提供。
type PaymentView struct {
	UserPaymentRecord
	OrderNo string `gorm:"column:order_no" json:"order_no"`
}
