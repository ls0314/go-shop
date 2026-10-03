package repository

import (
	"time"

	"demo-shop/services/marketing/internal/model"

	"gorm.io/gorm"
)

// UserCouponRepo 用户券表数据层实例
type UserCouponRepo struct {
	DB *gorm.DB
}

// NewUserCouponRepo 创建用户券表数据层实例
func NewUserCouponRepo(conn *gorm.DB) *UserCouponRepo {
	return &UserCouponRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (u *UserCouponRepo) WithTx(tx *gorm.DB) *UserCouponRepo {
	return &UserCouponRepo{DB: tx}
}

// CreateUserCoupon 领取落库,回填自增主键。
// status 走表默认值 unused,不在此显式写。
func (u *UserCouponRepo) CreateUserCoupon(coupon *model.UserCoupon) error {
	return u.DB.Create(coupon).Error
}

// CountUserCoupon 统计用户在指定模板下**占用限领名额**的券数。
// status != expired 才占名额:已过期的券不再占用额度。
// 必须在 LockCouponById 的锁内调用,配合行锁保证"读-判断-插入"串行化。
func (u *UserCouponRepo) CountUserCoupon(userId, templateId int64) (int64, error) {
	var cnt int64
	err := u.DB.Model(&model.UserCoupon{}).
		Where("user_id = ? AND template_id = ? AND status != ?", userId, templateId, model.CouponExpired).
		Count(&cnt).Error
	return cnt, err
}

// GetUserCouponList 分页查询某用户的券,联模板取名称/类型/门槛/优惠。
//
// 与"结算可用券"的差异:本方法**不过滤过期** —— 我的卡券页要展示已用/已过期。
// status 为空串表示不过滤。
func (u *UserCouponRepo) GetUserCouponList(userId int64, page, pageSize int, status string) ([]*model.UserCouponView, int64, error) {
	query := u.baseJoinQuery().Where("user_coupon.user_id = ?", userId)
	if status != "" {
		query = query.Where("user_coupon.status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.UserCouponView
	offset := (page - 1) * pageSize
	if err := query.
		Order("user_coupon.created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetAvailableCouponList 查用户全部可用券(结算选券用):未过期 + unused。
// 过滤条件与 UseCoupon 的核销条件保持一致 —— 两处口径若漂移,
// 会出现"结算页显示可用、下单却被拒"。
func (u *UserCouponRepo) GetAvailableCouponList(userId int64) ([]*model.UserCouponView, error) {
	var list []*model.UserCouponView
	err := u.baseJoinQuery().
		Where("user_coupon.expire_at > ? AND user_coupon.user_id = ? AND user_coupon.status = ?",
			time.Now(), userId, model.CouponUnused).
		Find(&list).Error
	return list, err
}

// GetUserCouponById 按ID查券(联模板),**不过滤状态与有效期**。
//
// 与 GetUsableCouponById 的分工:那个用于"核销前判定能不能用"(带
// status=unused + expire_at > now 条件),这个用于"核销后读回权威值"与
// "展示某张券当前是什么状态" —— 后两种场景券可能已经是 used/expired,
// 用带条件的查法会查不到。
func (u *UserCouponRepo) GetUserCouponById(userCouponId int64) (*model.UserCouponView, error) {
	var coupon model.UserCouponView
	err := u.baseJoinQuery().
		Where("user_coupon.user_coupon_id = ?", userCouponId).
		First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetUsableCouponById 取单张可用券(核销前校验用)。
// WHERE 含 expire_at > now 与 status = unused,查不到即视为不可用。
//
// 返回的 UserId 供调用方做归属校验 —— owner 不可变,无竞态。
func (u *UserCouponRepo) GetUsableCouponById(userCouponId int64) (*model.UserCouponView, error) {
	var coupon model.UserCouponView
	err := u.baseJoinQuery().
		Where("user_coupon.expire_at > ? AND user_coupon.user_coupon_id = ? AND user_coupon.status = ?",
			time.Now(), userCouponId, model.CouponUnused).
		First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetUserCouponByOrderNo 按订单号反查券(取消订单归还用)。
// 核销时写在 user_coupon.order_no 上,并有部分唯一索引兜底。
func (u *UserCouponRepo) GetUserCouponByOrderNo(orderNo string) (*model.UserCoupon, error) {
	var coupon model.UserCoupon
	if err := u.DB.Where("order_no = ?", orderNo).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// UseCoupon 核销:unused → used,写入 order_no / used_at。
//
// WHERE 含 status = unused + expire_at > now,并发下同一张券只有一个能成功;
// 返回 0 表示券不可用(不存在/已用/已过期)。
// 归属校验由调用方在调用前完成 —— owner 不可变,无竞态。
func (u *UserCouponRepo) UseCoupon(userCouponId int64, orderNo string) (int64, error) {
	res := u.DB.Model(&model.UserCoupon{}).
		Where("expire_at > ? AND user_coupon_id = ? AND status = ?",
			time.Now(), userCouponId, model.CouponUnused).
		Updates(map[string]interface{}{
			"status":   model.CouponUsed,
			"used_at":  time.Now(),
			"order_no": orderNo,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// ReturnCoupon 归还:used → unused,清空 order_no / used_at。
//
// WHERE status = used 与核销条件(status = unused)互斥,同一张券不可能同时被核销和归还;
// 重复归还时条件不匹配(返回 0),**天然幂等** —— 正是 Saga 补偿可重放所需要的。
func (u *UserCouponRepo) ReturnCoupon(userCouponId int64) (int64, error) {
	res := u.DB.Model(&model.UserCoupon{}).
		Where("user_coupon_id = ? AND status = ?", userCouponId, model.CouponUsed).
		Updates(map[string]interface{}{
			"status":   model.CouponUnused,
			"used_at":  nil,
			"order_no": nil,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// UpdateExpiredCoupon 把已过期但仍标 unused 的券置为 expired(对账用)。
//
// 单体没有这一步:它只靠查询条件 expire_at > now 过滤、状态不迁移。
// 拆出后状态与"限领名额"口径必须一致(CountUserCoupon 只看 status),
// 否则过期券会永久占用用户额度 —— 故补一个显式迁移。
func (u *UserCouponRepo) UpdateExpiredCoupon() (int64, error) {
	res := u.DB.Model(&model.UserCoupon{}).
		Where("status = ? AND expire_at <= ?", model.CouponUnused, time.Now()).
		Update("status", model.CouponExpired)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// baseJoinQuery 用户券 LEFT JOIN 模板:取券的模板信息(名称/类型/门槛/优惠)。
// 两张表都在 marketing_db,同库联表。
func (u *UserCouponRepo) baseJoinQuery() *gorm.DB {
	return u.DB.Model(&model.UserCoupon{}).
		Select("user_coupon.*, " +
			"coupon_template.coupon_name, " +
			"coupon_template.coupon_type, " +
			"coupon_template.threshold_amount, " +
			"coupon_template.discount_amount").
		Joins("LEFT JOIN coupon_template ON coupon_template.template_id = user_coupon.template_id")
}
