package model

import (
	"time"

	"gorm.io/datatypes"
)

type UserPayment struct {
	PaymentId int64          `gorm:"primary_key;autoIncrement;column:payment_id" json:"payment_id"`
	PayNo     string         `gorm:"column:pay_no" json:"pay_no"`
	OrderId   int64          `gorm:"column:order_id" json:"order_id"`
	UserId    int64          `gorm:"column:user_id" json:"user_id"`
	PayMethod string         `gorm:"column:pay_method" json:"pay_method"`
	PayAmount float64        `gorm:"column:pay_amount" json:"pay_amount"`
	PayStatus string         `gorm:"column:pay_status" json:"pay_status"`
	TradeNo   string         `gorm:"column:trade_no" json:"trade_no"`
	PayTime   time.Time      `gorm:"column:pay_time" json:"pay_time"`
	NotifyLog datatypes.JSON `gorm:"column:notify_log" json:"notify_log"`
	ExpireAt  time.Time      `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdateAdt time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (UserPayment) TableName() string { return "user_payment_record" }
