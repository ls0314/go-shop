package response

import "time"

type CreateCouponResp struct {
	TemplateId int64 `json:"template_id"`
}

type GetCouponList struct {
	TemplateId      int64     `gorm:"column:template_id" json:"template_id"`
	CouponName      string    `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string    `gorm:"column:coupon_type"  json:"coupon_type"`
	ThresholdAmount float64   `gorm:"column:threshold_amount"  json:"threshold_amount"`
	DiscountAmount  float64   `gorm:"column:discount_amount"  json:"discount_amount"`
	TotalCount      int64     `gorm:"column:total_count"  json:"total_count"`
	ReceivedCount   int64     `gorm:"column:received_count"  json:"received_count"`
	PerUserLimit    int64     `gorm:"column:per_user_limit"  json:"per_user_limit"`
	UsableDays      int64     `gorm:"column:usable_days"  json:"usable_days"`
	StartTime       time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime         time.Time `gorm:"column:end_time" json:"end_time"`
	Status          string    `gorm:"-" json:"status"`
}
type GetCouponListResp struct {
	List     []GetCouponList `json:"list"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int64           `json:"total"`
}

type UserReceiveCouponResp struct {
	UserCouponId int64     `json:"user_coupon_id"`
	ExpireTime   time.Time `json:"expire_time"`
}

type UserGetCouponList struct {
	UserCouponId    int64     `gorm:"column:user_coupon_id" json:"user_coupon_id"`
	CouponName      string    `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string    `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64   `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64   `gorm:"column:discount_amount" json:"discount_amount"`
	Status          string    `gorm:"column:status" json:"status"`
	ExpireAt        time.Time `gorm:"column:expire_at" json:"expire_at"`
	OrderNo         string    `gorm:"column:order_no" json:"order_no"`
	UsedAt          time.Time `gorm:"column:used_at" json:"used_at"`
	UserId          int64     `gorm:"column:user_id" json:"-"`
}

type UserGetCouponListResp struct {
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	List     []UserGetCouponList `json:"list"`
}

type GetAvailableCouponList struct {
	UserCouponId    int64   `gorm:"column:user_coupon_id" json:"user_coupon_id"`
	CouponName      string  `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string  `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64 `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64 `gorm:"column:discount_amount" json:"discount_amount"`
	PayAfter        float64 `gorm:"column:pay_after" json:"pay_after"`
}

type GetAvailableCouponResp struct {
	List []GetAvailableCouponList `json:"list"`
}
