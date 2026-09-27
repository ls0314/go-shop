package response

import "time"

// CreateCouponResp 创建优惠券模板响应（接口1）
type CreateCouponResp struct {
	TemplateId int64 `json:"template_id"`
}

// GetCouponList 优惠券模板列表项（接口2，管理端）
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

// UserReceiveCouponResp 用户领取优惠券响应（接口3）
type UserReceiveCouponResp struct {
	UserCouponId int64     `json:"user_coupon_id"`
	ExpireTime   time.Time `json:"expire_time"`
}

// UserGetCouponList 用户优惠券列表项（接口4）
// 由 user_coupon 表 LEFT JOIN coupon_template 组装；
// OrderNo 为核销时写入的订单号（未使用为空）；UserId 仅内部校验归属用，不返回前端
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

// UserGetCouponListResp 用户优惠券列表响应（接口4，分页）
type UserGetCouponListResp struct {
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	List     []UserGetCouponList `json:"list"`
}

// GetAvailableCouponList 结算可用券列表项（接口5）
// PayAfter 为使用该券后的实付金额（服务端计算），列表已按 PayAfter 升序排列
type GetAvailableCouponList struct {
	UserCouponId    int64   `gorm:"column:user_coupon_id" json:"user_coupon_id"`
	CouponName      string  `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string  `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64 `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64 `gorm:"column:discount_amount" json:"discount_amount"`
	PayAfter        float64 `gorm:"column:pay_after" json:"pay_after"`
}

// GetAvailableCouponResp 结算可用券响应（接口5）
type GetAvailableCouponResp struct {
	List []GetAvailableCouponList `json:"list"`
}

// UserCouponTemplate 领券中心模板列表项（用户可见）
// 相比管理端裁剪了 total_count/received_count 等运营字段；
// HeldCount 为当前用户在"未过期"口径下已领数量（服务端按 userId 子查询计算），
// 前端据此判断 held_count >= per_user_limit 时置灰"领取"按钮
type UserCouponTemplate struct {
	TemplateId      int64     `gorm:"column:template_id" json:"template_id"`
	CouponName      string    `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string    `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64   `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64   `gorm:"column:discount_amount" json:"discount_amount"`
	PerUserLimit    int64     `gorm:"column:per_user_limit" json:"per_user_limit"`
	HeldCount       int64     `gorm:"column:held_count" json:"held_count"`
	RemainingCount  int64     `gorm:"column:remaining_count" json:"remaining_count"`
	UsableDays      int64     `gorm:"column:usable_days" json:"usable_days"`
	StartTime       time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime         time.Time `gorm:"column:end_time" json:"end_time"`
}

// UserCouponTemplateListResp 领券中心模板列表响应（分页）
type UserCouponTemplateListResp struct {
	List     []UserCouponTemplate `json:"list"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Total    int64                `json:"total"`
}
