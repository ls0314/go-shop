package model

import "time"

// CouponTemplate 优惠券模板实体（coupon_template 表）
// 对应管理端创建的优惠券定义：类型/门槛/优惠力度/发放总量/有效期模式
// 有效期模式二选一：UsableDays>0 相对有效期；否则用 StartTime/EndTime 固定有效期
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

// UserCoupon 用户优惠券实体（user_coupon 表）
// 用户领取模板后生成的券实例：状态机 unused → used（核销）/ expired（过期），
// 取消订单可 used → unused（归还）。UsedAt 用指针类型区分"未使用(NULL)"与零值
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
