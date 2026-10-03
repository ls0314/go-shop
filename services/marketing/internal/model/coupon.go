package model

import "time"

// CouponTemplate 优惠券模板结构体, 对应数据表coupon_template
//
// 有效期模式二选一:
//   - UsableDays > 0 → 领取后 N 天有效(相对)
//   - 否则用 StartTime / EndTime(固定)
type CouponTemplate struct {
	TemplateId      int64     `gorm:"primaryKey;column:template_id" json:"template_id"`
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

func (CouponTemplate) TableName() string {
	return "coupon_template"
}

// UserCoupon 用户优惠券结构体, 对应数据表user_coupon
//
// 状态机:unused → used(核销) / unused → expired(过期);used → unused(取消订单归还)。
// UsedAt 用指针区分"未使用(NULL)"与零值。
type UserCoupon struct {
	UserCouponId int64      `gorm:"primaryKey;column:user_coupon_id" json:"user_coupon_id"`
	TemplateId   int64      `gorm:"column:template_id" json:"template_id"`
	UserId       int64      `gorm:"column:user_id" json:"user_id"`
	Status       string     `gorm:"column:status" json:"status"`
	OrderNo      string     `gorm:"column:order_no" json:"order_no"`
	UsedAt       *time.Time `gorm:"column:used_at" json:"used_at"`
	ExpireAt     time.Time  `gorm:"column:expire_at" json:"expire_at"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
}

func (UserCoupon) TableName() string {
	return "user_coupon"
}

// UserCouponView 用户券 + 联表取到的模板快照。
//
// 为什么不把四个模板字段塞进 UserCoupon:那些列不在 user_coupon 表上,
// 塞进实体会让"实体 ↔ 表"的对应失真 —— 一旦有人拿它去 Create/Update,
// GORM 会尝试写不存在的列。
type UserCouponView struct {
	UserCoupon
	CouponName      string  `gorm:"column:coupon_name" json:"coupon_name"`
	CouponType      string  `gorm:"column:coupon_type" json:"coupon_type"`
	ThresholdAmount float64 `gorm:"column:threshold_amount" json:"threshold_amount"`
	DiscountAmount  float64 `gorm:"column:discount_amount" json:"discount_amount"`
}

// CouponForUser 领券中心的模板项:模板 + 该用户已领数与剩余量
type CouponForUser struct {
	CouponTemplate
	HeldCount      int64 `gorm:"column:held_count" json:"held_count"`
	RemainingCount int64 `gorm:"column:remaining_count" json:"remaining_count"`
}

// 券状态,与 user_coupon.status 的 CHECK 约束一致
const (
	CouponUnused  = "unused"
	CouponUsed    = "used"
	CouponExpired = "expired"
)

// 券类型,与 coupon_template.coupon_type 的 CHECK 约束一致
const (
	CouponTypeFullReduction  = "full_reduction"
	CouponTypeDirectDiscount = "direct_discount"
)

// 模板展示态:按当前时间推导,不是库里的列。
// 由服务端算而非前端算 —— 两端各自判时间会因时钟漂移给出不同结论。
const (
	CouponStatusNotStarted = "not_started"
	CouponStatusOngoing    = "ongoing"
	CouponStatusEnded      = "ended"
)

// DeriveCouponStatus 按有效期推导模板展示态。
//
// 相对有效期(UsableDays > 0)的模板**不判结束**:券是领取后才开始计时的,
// 模板本身没有"过期"一说。这与领券中心 ListCouponForUser 的过滤口径
// (`usable_days > 0 OR end_time > now()`)一致。
func DeriveCouponStatus(t *CouponTemplate, now time.Time) string {
	if t == nil || t.IsDeleted {
		return CouponStatusEnded
	}
	if t.UsableDays > 0 {
		return CouponStatusOngoing
	}
	if !t.StartTime.IsZero() && now.Before(t.StartTime) {
		return CouponStatusNotStarted
	}
	if !t.EndTime.IsZero() && now.After(t.EndTime) {
		return CouponStatusEnded
	}
	return CouponStatusOngoing
}

// CalcPayAmount 算用券后的实付金额。满减 = 金额 - 优惠;直减 = 金额 × 折扣率。
// 不满足门槛时返回 (0, false)。
//
// 放在券域而不是调用方:这是券的语义,不是订单的。单体时代下单链路里
// 复制过一份同样的算法(order_service.go),拆出后由本服务算好返回,
// 调用方不再需要知道满减与直减的区别。
func CalcPayAmount(couponType string, orderAmount, thresholdAmount, discountAmount float64) (float64, bool) {
	if orderAmount < thresholdAmount {
		return 0, false
	}
	if couponType == CouponTypeFullReduction {
		return orderAmount - discountAmount, true
	}
	return orderAmount * discountAmount, true
}
