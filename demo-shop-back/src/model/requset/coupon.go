package requset

import "time"

type CreateCouponReq struct {
	CouponName      string    `json:"coupon_name"`
	CouponType      string    `json:"coupon_type"`
	ThresholdAmount float64   `json:"threshold_amount"`
	DiscountAmount  float64   `json:"discount_amount"`
	TotalCount      int64     `json:"total_count"`
	PerUserLimit    int64     `json:"per_user_limit"`
	UsableDays      int64     `json:"usable_days"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
}

type GetCouponListReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	CouponName string `form:"coupon_name"`
	CouponType string `form:"coupon_type"`
}

type UserGetCouponReq struct {
	TemplateId int64 `form:"template_id"`
}

type UserGetCouponListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

type GetAvailableCouponReq struct {
	OrderAmount float64 `form:"order_amount"`
}
