package requset

import "time"

// CreateCouponReq 创建优惠券模板请求体（接口1）
// 有效期二选一：UsableDays>0 相对有效期；UsableDays=0 时必须传 StartTime/EndTime 固定有效期
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

// GetCouponListReq 优惠券模板列表查询参数（接口2，管理端）
type GetCouponListReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	CouponName string `form:"coupon_name"`
	CouponType string `form:"coupon_type"`
}

// UserGetCouponReq 按模板查询用户券参数（预留，暂未使用）
type UserGetCouponReq struct {
	TemplateId int64 `form:"template_id"`
}

// UserGetCouponListReq 用户优惠券列表查询参数（接口4）
// Status 支持 unused/used/expired，为空返回全部状态
type UserGetCouponListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

// GetAvailableCouponReq 结算可用券查询参数（接口5）
// OrderAmount 为订单总金额（未抵扣前），小于 0 返回 11009
type GetAvailableCouponReq struct {
	OrderAmount float64 `form:"order_amount"`
}

// UserGetTemplateListReq 领券中心模板列表查询参数
// 用户端展示可领取的券模板,分页结构与其他列表接口保持一致
type UserGetTemplateListReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
