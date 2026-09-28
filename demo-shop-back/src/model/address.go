package model

import "time"

type UserAddress struct {
	AddressId     int64     `gorm:"primary_key;autoIncrement;column:address_id" json:"address_id"`
	UserId        int64     `gorm:"column:user_id" json:"user_id"`
	ReceiverName  string    `gorm:"column:receiver_name" json:"receiver_name"`
	ReceiverPhone string    `gorm:"column:receiver_phone" json:"receiver_phone"`
	Province      string    `gorm:"column:province" json:"province"`
	City          string    `gorm:"column:city" json:"city"`
	District      string    `gorm:"column:district" json:"district"`
	DetailAddress string    `gorm:"column:detail_address" json:"detail_address"`
	PostalCode    string    `gorm:"column:postal_code" json:"postal_code"`
	IsDefault     bool      `gorm:"column:is_default" json:"is_default"`
	AddressTag    string    `gorm:"column:address_tag" json:"address_tag"`
	IsDeleted     bool      `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserAddress) TableName() string { return "user_address" }
