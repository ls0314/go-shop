package repository

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ============================================================
// 券域孤儿方法 —— **C4 时整文件删除**
// ============================================================
//
// 这些方法在 C3 之前各有调用方;券域的读接口与领券改经 RPC 后,
// 单体已 **0 调用方**(2025 复查:逐个方法在全仓 `src/` 与 `tests/` 内检索确认)。
//
// 为什么拆到单独文件而不是直接删:
//   - 它们与仍在下单链路使用的四个方法原本同处 coupon_repo.go。直接删要
//     逐方法判断"这个还有没有人用",判错一次就是编译失败或悄悄换实现;
//     拆成"在用/孤儿"两个文件后,**读的人一眼能看出边界**;
//   - C4 把下单链路也改成 RPC 之后,coupon.go 里那四个方法会一起消失,
//     届时本文件与 coupon.go 一并删除,不需要再翻一遍。
//
// 为什么留着而不是现在就删:
//   它们承载着券域的**实现细节与既有注释**(双防线怎么做的、状态机怎么约束的),
//   而 services/marketing 的实现正是照此平移的。留到 C4 能作为对照,
//   确认搬迁无遗漏 —— 删早了就少了这份参照。
//
// 对应实现已在:
//
//	services/marketing/internal/repository/coupon.go      (模板)
//	services/marketing/internal/repository/usercoupon.go   (用户券)
//	services/marketing/internal/infra/gate/coupongate.go   (领券闸门 Lua)

// CreateCoupon 创建优惠券模板
// 接收值：coupon - 优惠券模板实体（coupon_name/coupon_type/金额/数量/有效期等）
// 返回值：template_id - 新建模板的主键ID
//
//	err - 错误信息，创建失败时返回
func (c *CouponRepo) CreateCoupon(coupon model.CouponTemplate) (int64, error) {
	err := c.db.Create(&coupon).Error
	return coupon.TemplateId, err
}

// GetCouponList 分页查询优惠券模板列表（管理端）
// 接收值：req - 查询条件（page/page_size/coupon_name 精确匹配/coupon_type 筛选）
// 返回值：[]response.GetCouponList - 模板列表（已过滤 is_deleted=false，按 created_at 倒序）
//
//	total - 符合条件的模板总数
//	err - 错误信息
func (c *CouponRepo) GetCouponList(req *requset.GetCouponListReq) ([]response.GetCouponList, int64, error) {
	var coupons []response.GetCouponList

	// 基础查询：仅查未删除（软删除）的模板
	baseQuery := c.db.Model(&model.CouponTemplate{}).
		Where("is_deleted = ?", false)

	// 可选筛选条件：名称/类型，仅在传入时追加 WHERE
	if req.CouponName != "" {
		baseQuery = baseQuery.Where("coupon_name = ?", req.CouponName)
	}
	if req.CouponType != "" {
		baseQuery = baseQuery.Where("coupon_type = ?", req.CouponType)
	}

	// 先统计总数（分页响应需要），再查询当前页数据——两次查询共享同一 baseQuery 条件
	var total int64
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize
	// 分页 + 按创建时间倒序（最新创建的模板排最前）
	err = baseQuery.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&coupons).Error
	if err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// GetTemplateById 普通读取模板(无锁、无事务要求)
func (c *CouponRepo) GetTemplateById(templateId int64) (*model.CouponTemplate, error) {
	var tpl model.CouponTemplate
	err := c.db.Where("template_id = ?", templateId).First(&tpl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrCouponTemplateNotExist
		}
		return nil, err
	}
	return &tpl, nil
}

// GetAllActiveTemplates 全量读取未删除模板
//
// 曾是券闸门对账的数据源。该任务已于 C3 删除(见 src/task/init_recocile.go):
// 它读本库那张空表会把 coupon:stock:* 收敛成 0,反而打死 marketing-service 的领券。
func (c *CouponRepo) GetAllActiveTemplates() ([]model.CouponTemplate, error) {
	var templateList []model.CouponTemplate
	err := c.db.Where("is_deleted = ?", false).Find(&templateList).Error
	return templateList, err
}

// LockTemplateByID 事务内以 FOR UPDATE 锁定模板行，并读取锁内最新数据
// 接收值：templateId - 优惠券模板ID
// 返回值：*model.CouponTemplate - 锁内读取的模板最新数据
//
//	err - 模板不存在返回 ErrCouponTemplateNotExist
//
// 说明：必须由事务内的 CouponRepo 调用（WithTx 之后），否则语句结束锁即释放；
//
//	锁的生命周期 = 事务生命周期，用于串行化同一模板的并发领取 —— 双防线的第一道
func (c *CouponRepo) LockTemplateByID(templateId int64) (*model.CouponTemplate, error) {
	var tpl model.CouponTemplate
	// clause.Locking{Strength: "UPDATE"} 生成 SELECT ... FOR UPDATE：
	// 锁从本语句开始持有，直到事务 commit/rollback 才释放（悲观锁）
	err := c.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("template_id = ?", templateId).
		First(&tpl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrCouponTemplateNotExist
		}
		return nil, err
	}
	return &tpl, nil
}

// ReceiveCoupon 原子条件扣减领取余量（乐观锁，防超发）—— 双防线的第二道
// 接收值：templateId - 模板ID；totalCount - 发放总量
// 返回值：rowsAffected - 影响行数，0 表示余量已空
//
//	err - 错误信息
//
// 说明：WHERE received_count < total_count 保证并发下已领数永不超总量
func (c *CouponRepo) ReceiveCoupon(templateId, totalCount int64) (int64, error) {
	res := c.db.Model(&model.CouponTemplate{}).
		Where("template_id = ? AND received_count < ?", templateId, totalCount).
		Updates(map[string]interface{}{
			"received_count": gorm.Expr("received_count + 1"),
		})
	if res.Error != nil {
		return -1, res.Error
	}
	return res.RowsAffected, nil
}

// CreateUserCoupon 领取落库，回填主键
func (c *CouponRepo) CreateUserCoupon(userCoupon model.UserCoupon) (int64, error) {
	err := c.db.Create(&userCoupon).Error
	return userCoupon.UserCouponId, err
}

// CountUserCoupon 统计用户已领取某模板的券数（限领校验用）
// 说明：status != 'expired' 才占名额——已过期的券不再占用额度
func (c *CouponRepo) CountUserCoupon(userId, templateId int64) (int64, error) {
	var count int64
	err := c.db.Model(&model.UserCoupon{}).
		Where("user_id = ? AND template_id = ? AND status != ?", userId, templateId, "expired").
		Count(&count).Error
	return count, err
}

// GetUserCouponList 查询用户全部券（无分页，结算选券用）
// 说明：过滤未过期 + 未使用
func (c *CouponRepo) GetUserCouponList(userId int64) ([]response.UserGetCouponList, error) {
	var coupons []response.UserGetCouponList
	err := c.db.Model(&model.UserCoupon{}).
		Select("user_coupon.user_coupon_id, user_coupon.status, user_coupon.order_no, "+
			"user_coupon.expire_at, coupon_template.coupon_name, coupon_template.coupon_type, "+
			"coupon_template.threshold_amount, coupon_template.discount_amount").
		Joins("LEFT JOIN coupon_template ON coupon_template.template_id = user_coupon.template_id").
		Where("user_coupon.expire_at > ? AND user_coupon.user_id = ? AND user_coupon.status = ?",
			time.Now(), userId, model.CouponUnused).
		Find(&coupons).Error
	return coupons, err
}

// UserGetCouponList 分页查询用户券（我的卡券，不过滤过期）
func (c *CouponRepo) UserGetCouponList(userId int64, req *requset.UserGetCouponListReq) ([]response.UserGetCouponList, int64, error) {
	var coupons []response.UserGetCouponList
	var total int64

	baseQuery := c.db.Model(&model.UserCoupon{}).
		Joins("LEFT JOIN coupon_template ON coupon_template.template_id = user_coupon.template_id").
		Where("user_coupon.user_id = ?", userId)
	if req.Status != "" {
		baseQuery = baseQuery.Where("user_coupon.status = ?", req.Status)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := baseQuery.
		Select("user_coupon.user_coupon_id, user_coupon.status, user_coupon.order_no, " +
			"user_coupon.used_at, user_coupon.expire_at, coupon_template.coupon_name, " +
			"coupon_template.coupon_type, coupon_template.threshold_amount, coupon_template.discount_amount").
		Order("user_coupon.created_at DESC").
		Limit(req.PageSize).Offset(offset).
		Find(&coupons).Error
	return coupons, total, err
}

// UserGetTemplateList 领券中心分页查询（含该用户已领数与剩余量）
func (c *CouponRepo) UserGetTemplateList(userId int64, page, pageSize int) ([]response.UserCouponTemplate, int64, error) {
	var total int64
	baseQuery := c.db.Model(&model.CouponTemplate{}).
		Where("is_deleted = ?", false).
		Where("usable_days > 0 OR end_time > ?", time.Now())

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// held_count 子查询:与 CountUserCoupon 的口径必须一致
	heldCountSub := c.db.Model(&model.UserCoupon{}).
		Select("COUNT(*)").
		Where("template_id = coupon_template.template_id AND user_id = ? AND status != ?",
			userId, "expired")

	var list []response.UserCouponTemplate
	err := baseQuery.
		Select("coupon_template.template_id, coupon_template.coupon_name, coupon_template.coupon_type, "+
			"coupon_template.threshold_amount, coupon_template.discount_amount, "+
			"coupon_template.per_user_limit, coupon_template.usable_days, "+
			"coupon_template.start_time, coupon_template.end_time, "+
			"coupon_template.total_count - coupon_template.received_count AS remaining_count, "+
			"(?) AS held_count", heldCountSub).
		Order("coupon_template.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Scan(&list).Error
	return list, total, err
}
