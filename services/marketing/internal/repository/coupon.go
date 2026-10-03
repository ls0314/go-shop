package repository

import (
	"demo-shop/services/marketing/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CouponRepo 优惠券模板表数据层实例
type CouponRepo struct {
	DB *gorm.DB
}

// NewCouponRepo 创建优惠券模板表数据层实例
func NewCouponRepo(conn *gorm.DB) *CouponRepo {
	return &CouponRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (c *CouponRepo) WithTx(tx *gorm.DB) *CouponRepo {
	return &CouponRepo{DB: tx}
}

// CreateCoupon 创建优惠券模板,回填自增主键
func (c *CouponRepo) CreateCoupon(coupon *model.CouponTemplate) error {
	return c.DB.Create(coupon).Error
}

// GetCouponById 根据ID查询模板。
// 查不到直接返回 gorm.ErrRecordNotFound,由调用方判定为"模板不存在"。
func (c *CouponRepo) GetCouponById(templateId int64) (*model.CouponTemplate, error) {
	var coupon model.CouponTemplate
	if err := c.DB.Where("template_id = ?", templateId).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// GetCouponList 分页查询模板(已过滤 is_deleted,按 created_at 倒序)。
// coupon_name 精确匹配、coupon_type 枚举筛选,空串表示不过滤 —— 与单体一致(非模糊搜索)。
func (c *CouponRepo) GetCouponList(page, pageSize int, couponName, couponType string) ([]*model.CouponTemplate, int64, error) {
	var coupons []*model.CouponTemplate
	var total int64

	query := c.DB.Model(&model.CouponTemplate{}).
		Where("is_deleted = ?", false).
		Order("created_at DESC")
	if couponName != "" {
		query = query.Where("coupon_name = ?", couponName)
	}
	if couponType != "" {
		query = query.Where("coupon_type = ?", couponType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// GetActiveCouponList 全量读取未删除模板(闸门对账用)
func (c *CouponRepo) GetActiveCouponList() ([]*model.CouponTemplate, error) {
	var coupons []*model.CouponTemplate
	err := c.DB.Where("is_deleted = ?", false).Find(&coupons).Error
	return coupons, err
}

// LockCouponById 事务内以 FOR UPDATE 锁定模板行并读取锁内最新数据。
//
// 必须由事务内的实例调用(WithTx 之后),否则语句结束锁即释放。
// 锁的生命周期 = 事务生命周期,用于串行化同一模板的并发领取 —— 双防线的第一道。
func (c *CouponRepo) LockCouponById(templateId int64) (*model.CouponTemplate, error) {
	var coupon model.CouponTemplate
	err := c.DB.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("template_id = ?", templateId).
		First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// UpdateCoupon 全量更新模板(局部更新由调用方合并后传入)
func (c *CouponRepo) UpdateCoupon(coupon *model.CouponTemplate) error {
	return c.DB.Save(coupon).Error
}

// DeleteCoupon 软删除模板
func (c *CouponRepo) DeleteCoupon(templateId int64) error {
	return c.DB.Model(&model.CouponTemplate{}).
		Where("template_id = ?", templateId).
		Update("is_deleted", true).Error
}

// IncrReceivedCount 原子条件扣减领取余量(乐观锁,防超发)—— 双防线的第二道。
//
// WHERE received_count < total_count 保证并发下已领数永不超总量;
// 返回 0 表示余量已空,由调用方判定为"已领完"。
func (c *CouponRepo) IncrReceivedCount(templateId, totalCount int64) (int64, error) {
	res := c.DB.Model(&model.CouponTemplate{}).
		Where("template_id = ? AND received_count < ?", templateId, totalCount).
		Update("received_count", gorm.Expr("received_count + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DecrReceivedCount 归还领取计数(对账在不一致时纠偏用)。
// 领取路径本身在事务里,回滚天然撤销 +1,不需要它。
func (c *CouponRepo) DecrReceivedCount(templateId int64) error {
	return c.DB.Model(&model.CouponTemplate{}).
		Where("template_id = ? AND received_count > 0", templateId).
		Update("received_count", gorm.Expr("received_count - 1")).Error
}

// 表名常量。领券中心的查询不能走 Model(&X{}) —— GORM 会给它起别名
// ("coupon_templates"),而 held_count 子查询里引用的 coupon_template.template_id
// 就指不到真表了。product-service 的关联查询同样用 Table(真实表名) 规避。
const (
	couponTemplateTable = "coupon_template"
	userCouponTable     = "user_coupon"
)

// ListCouponForUser 领券中心分页查询:模板 + 该用户已领数与剩余量。
//
// 与管理端 GetCouponList 的差异:
//   - 有效期过滤:usable_days > 0(相对,领取后才计时)OR end_time > now(固定未结束)
//   - held_count 只算"未过期"的券,与 CountUserCoupon 同口径
func (c *CouponRepo) ListCouponForUser(userId int64, page, pageSize int) ([]*model.CouponForUser, int64, error) {
	query := c.DB.Table(couponTemplateTable).
		Where(couponTemplateTable+".is_deleted = ?", false).
		Where(couponTemplateTable + ".usable_days > 0 OR " + couponTemplateTable + ".end_time > NOW()")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 与 CountUserCoupon 的口径必须一致(status != 'expired'),否则"已领数"与
	// "限领判断"会给出不同结论 —— 前端置灰的按钮点下去仍能领到
	heldSub := c.DB.Table(userCouponTable).
		Select("COUNT(*)").
		Where(userCouponTable+".template_id = "+couponTemplateTable+".template_id").
		Where(userCouponTable+".user_id = ? AND "+userCouponTable+".status != ?", userId, model.CouponExpired)

	var list []*model.CouponForUser
	err := query.
		Select(couponTemplateTable+".*, "+
			couponTemplateTable+".total_count - "+couponTemplateTable+".received_count AS remaining_count, "+
			"(?) AS held_count", heldSub).
		Order(couponTemplateTable + ".created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}
