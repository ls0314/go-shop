package model

import "time"

type CouponTemplate struct {
	TemplateId      int64     `gorm:"column:template_id;primary_key;AUTO_INCREMENT " json:"template_id"`
	CouponName      string    `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string    `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64   `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64   `gorm:"column:discount_amount" json:"discount_amount"`
	TotalCount      int64     `gorm:"column:total_count" json:"total_count"`
	ReceivedCount   int64     `gorm:"column:received_count" json:"received_count"`
	PerUserLimit    int64     `gorm:"column:per_user_limit" json:"per_user_limit"`
	UsableDays      int64     `gorm:"column:usable_days" json:"usable_days"`
	StartTime       time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime         time.Time `gorm:"column:end_time" json:"end_time"`
	IsDeleted       bool      `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CouponTemplate) TableName() string { return "coupon_template" }

type UserCoupon struct {
	UserCouponId int64      `gorm:"column:user_coupon_id;primary_key;AUTO_INCREMENT" json:"user_coupon_id"`
	TemplateId   int64      `gorm:"column:template_id" json:"template_id"`
	UserId       int64      `gorm:"column:user_id" json:"user_id"`
	Status       string     `gorm:"column:status" json:"status"`
	OrderNo      string     `gorm:"column:order_no" json:"order_no"`
	UsedAt       *time.Time `gorm:"column:used_at" json:"used_at"`
	ExpireAt     time.Time  `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
}

func (UserCoupon) TableName() string { return "user_coupon" }
