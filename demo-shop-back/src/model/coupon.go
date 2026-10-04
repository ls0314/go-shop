package model

import "time"

// ============================================================
// 券域实体的**过渡期形态**
// ============================================================
//
// 券表(coupon_template / user_coupon)的所有权已迁 marketing-service 的
// 独立库 marketing_db(C3)。本文件里的两个结构体**不再是落库实体** ——
// 本进程已没有任何代码读写那两张表(repository/coupon*.go 已删)。
//
// 它们现在只服务一件事:infra/couponclient 把 RPC 返回的 proto 转成
// **本进程的响应形状**,而 HTTP 响应的 JSON 字段名是对前端的契约 ——
// 不能因为换了实现就改名。所以这里保留 json tag 与字段名。
//
// **gorm tag 已经是"化石"**:它们记录的是迁移前 demo_shop 里的列名,
// 留着只为对照(排查"这个字段以前叫什么"时有用),不代表还能用它建表。
// 下面两个 TableName() 同理 —— 没有任何调用方,但删掉会让
// "这些字段曾经对应哪张表"这条线索断掉。
// 等 BFF 直接返回 proto 时(它不需要这层转换),本文件可整体删除。
// ============================================================

// CouponTemplate 优惠券模板(管理端创建的券定义)。
//
// 有效期模式二选一:UsableDays>0 相对有效期;否则用 StartTime/EndTime
// 固定有效期。这个"二选一"的校验在 marketing-service 侧
// (与 coupon_template 的 CHECK 约束是一套规则,拆开维护必然漂移)。
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

// UserCoupon 用户领取模板后生成的券实例。
//
// 状态机 unused → used(核销)/ expired(过期),取消订单可 used → unused
// (归还)。UsedAt 用指针类型区分"未使用(NULL)"与零值。
//
// **注意本结构体没有 idempotency_key**:那是 C4 给 marketing_db 加的新列
// (幂等键改由调用方传入),只存在于服务端。本进程做 RPC → HTTP 转换时
// 不需要它 —— 幂等键是服务间的事,不该出现在给前端的响应里。
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
