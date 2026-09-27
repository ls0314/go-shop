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
// 优惠券模块数据层定义及实例化部分
// ============================================================

// CouponRepo 优惠券模块数据层实例
type CouponRepo struct {
	db *gorm.DB
}

// NewCouponRepo 新建优惠券模块数据层实例
// 接收值：无接收值，使用全局数据库连接
// 返回值：*CouponRepo - 优惠券模块数据层实例指针
func NewCouponRepo(conn *gorm.DB) *CouponRepo {
	return &CouponRepo{db: conn}
}

// WithTx 优惠券表事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*CouponRepo - 绑定事务的优惠券模块数据层指针
// 说明：事务内所有方法必须使用本方法返回的实例，保证读写同属一个事务
func (c *CouponRepo) WithTx(tx *gorm.DB) *CouponRepo {
	return &CouponRepo{
		db: tx,
	}
}

// ============================================================
// 优惠券模板相关
// ============================================================

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
//	锁的生命周期 = 事务生命周期，用于串行化同一模板的并发领取
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

// ReceiveCoupon 原子条件扣减模板领取余量（乐观锁，防超发）
// 接收值：templateId - 优惠券模板ID
//
//	totalCount - 模板发放总量（作为扣减上限条件）
//
// 返回值：rowsAffected - 影响行数，0 表示余量已空（received_count >= total_count）
//
//	err - 错误信息
//
// 说明：WHERE received_count < total_count 保证并发下已领取数永远不会超过总量
func (c *CouponRepo) ReceiveCoupon(templateId, totalCount int64) (int64, error) {
	// 条件扣减（乐观锁）：WHERE received_count < total_count 保证并发下不超发；
	// gorm.Expr 生成 SQL 表达式 received_count + 1，在数据库侧原子自增
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

// ============================================================
// 用户优惠券相关
// ============================================================

// CreateUserCoupon 创建用户优惠券（领取落库）
// 接收值：userCoupon - 用户券实体（template_id/user_id/expire_at，status 走表默认值 unused）
// 返回值：user_coupon_id - 新建用户券的主键ID
//
//	err - 错误信息
func (c *CouponRepo) CreateUserCoupon(userCoupon model.UserCoupon) (int64, error) {
	err := c.db.Create(&userCoupon).Error
	return userCoupon.UserCouponId, err
}

// CountUserCoupon 统计用户在指定模板下占用限领名额的券数
// 接收值：userId - 用户ID；templateId - 模板ID
// 返回值：count - 已领取且未过期的券数量（status != 'expired' 不占名额）
//
//	err - 错误信息
//
// 说明：必须在 LockTemplateByID 的锁内调用，配合行锁保证"读-判断-插入"串行化
func (c *CouponRepo) CountUserCoupon(userId, templateId int64) (int64, error) {
	var cnt int64
	err := c.db.Model(&model.UserCoupon{}).
		Where("user_id = ? AND template_id = ? AND status != ?", userId, templateId, "expired").
		Count(&cnt).Error
	return cnt, err
}

// GetUserCouponList 查询用户全部可用券（结算选券用）
// 接收值：userId - 用户ID
// 返回值：[]response.UserGetCouponList - 该用户未过期且 status='unused' 的券（JOIN 模板取名称/类型/金额信息）
//
//	err - 错误信息
//
// 说明：接口5（结算可用券）的数据源，过滤条件与 UseCoupon 核销条件保持一致
func (c *CouponRepo) GetUserCouponList(userId int64) ([]response.UserGetCouponList, error) {
	var coupons []response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
	// 主表 user_coupon LEFT JOIN coupon_template：取券的模板信息（名称/类型/门槛/优惠）
	// WHERE 三重过滤：未过期 + 属于该用户 + 未使用——与 UseCoupon 核销条件保持一致
	baseQuery := c.db.Model(&model.UserCoupon{}).
		Select(templateTable+".coupon_name,"+
			templateTable+".coupon_type,"+
			templateTable+".threshold_amount,"+
			templateTable+".discount_amount,"+
			userCouponTable+".user_coupon_id,"+
			userCouponTable+".status,"+
			userCouponTable+".order_no,"+
			userCouponTable+".used_at,"+
			userCouponTable+".expire_at").
		Where(userCouponTable+".expire_at > ? AND "+userCouponTable+".user_id = ? AND "+userCouponTable+".status = ? ", time.Now(), userId, "unused").
		Joins("LEFT JOIN " + templateTable + " ON " + templateTable + ".template_id = " + userCouponTable + ".template_id")
	err := baseQuery.Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetUserCoupon 查询单张用户券（下单核销前使用，含归属校验所需 user_id）
// 接收值：userCouponId - 用户券ID
// 返回值：*response.UserGetCouponList - 该券详情（含模板信息与 user_id）
//
//	err - 券不存在/已使用/已过期返回 ErrCouponNotExistOrUsed
//
// 说明：WHERE 含 expire_at > now() 与 status='unused'，查询不到即视为不可用；
//
//	与 UseCoupon 的更新条件保持完全一致（乐观锁铁律）
func (c *CouponRepo) GetUserCoupon(userCouponId int64) (*response.UserGetCouponList, error) {
	var coupon response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
	// 单券查询：JOIN 模板取优惠信息，额外 SELECT user_id 供服务层做归属校验（防越权核销）
	baseQuery := c.db.Model(&model.UserCoupon{}).
		Select(templateTable+".coupon_name,"+
			templateTable+".coupon_type,"+
			templateTable+".threshold_amount,"+
			templateTable+".discount_amount,"+
			userCouponTable+".user_coupon_id,"+
			userCouponTable+".status,"+
			userCouponTable+".order_no,"+
			userCouponTable+".used_at,"+
			userCouponTable+".user_id,"+
			userCouponTable+".expire_at").
		Where(userCouponTable+".expire_at > ? AND "+userCouponTable+".user_coupon_id = ? AND "+userCouponTable+".status = ? ", time.Now(), userCouponId, "unused").
		Joins("LEFT JOIN " + templateTable + " ON " + templateTable + ".template_id = " + userCouponTable + ".template_id")
	err := baseQuery.First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrCouponNotExistOrUsed
		}
		return nil, err
	}
	return &coupon, nil
}

// UserGetCouponList 分页查询当前用户的优惠券列表（接口4）
// 接收值：userId - 用户ID；req - 分页与状态筛选（status 为空返回全部）
// 返回值：[]response.UserGetCouponList - 用户券列表（JOIN 模板信息，按 created_at 倒序）
//
//	total - 总数
//	err - 错误信息
//
// 说明：不过滤过期券——"我的卡券"页需要展示 expired 状态；与结算选券（GetUserCouponList）语义不同
func (c *CouponRepo) UserGetCouponList(userId int64, req *requset.UserGetCouponListReq) ([]response.UserGetCouponList, int64, error) {
	var coupons []response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
	// 接口4数据源：仅按 user_id 过滤（含全部状态），与结算选券（GetUserCouponList）不同——
	// "我的卡券"页需要展示已用/已过期券
	baseQuery := c.db.Model(&model.UserCoupon{}).
		Select(templateTable+".coupon_name,"+
			templateTable+".coupon_type,"+
			templateTable+".threshold_amount,"+
			templateTable+".discount_amount,"+
			userCouponTable+".user_coupon_id,"+
			userCouponTable+".status,"+
			userCouponTable+".order_no,"+
			userCouponTable+".used_at,"+
			userCouponTable+".expire_at").
		Where(userCouponTable+".user_id = ?", userId).
		Joins("LEFT JOIN " + templateTable + " ON " + templateTable + ".template_id = " + userCouponTable + ".template_id")

	if req.Status != "" {
		baseQuery = baseQuery.Where(userCouponTable+".status = ?", req.Status)
	}

	var total int64
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize
	err = baseQuery.Offset(offset).Limit(pageSize).Order(userCouponTable + ".created_at DESC").Find(&coupons).Error
	if err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// GetUserCouponByOrderNo 按订单号反查用户券（取消订单归还时使用）
// 接收值：orderNo - 订单号（核销时已写入 user_coupon.order_no）
// 返回值：*model.UserCoupon - 该订单使用的用户券
//
//	err - 无匹配记录返回 ErrCouponNotExistOrUsed
func (c *CouponRepo) GetUserCouponByOrderNo(orderNo string) (*model.UserCoupon, error) {
	var coupon model.UserCoupon
	err := c.db.Model(model.UserCoupon{}).
		Where("order_no = ?", orderNo).
		First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrCouponNotExistOrUsed
		}
		return nil, err
	}
	return &coupon, nil
}

// ============================================================
// 领券中心：用户可见模板列表（用户端独立查询，不复用管理端 GetCouponList）
// ============================================================

// UserGetTemplateList 领券中心模板分页查询（用户端）
// 接收值：userId - 当前用户ID（用于子查询计算 held_count）；page/pageSize - 分页
// 返回值：[]response.UserCouponTemplate - 模板列表；total - 总数；err
//
// 过滤条件（与管理端 GetCouponList 的差异）：
//
//	① is_deleted = false 同管理端
//	② 有效期过滤：usable_days > 0（相对有效期，领取后才计时，永不过期）OR end_time > now()（固定有效期未结束）
//	③ held_count 子查询：与 CountUserCoupon 口径一致（status != 'expired' 不占名额）
//	④ remaining_count = total_count - received_count 由 SQL 表达式计算
func (c *CouponRepo) UserGetTemplateList(userId int64, page, pageSize int) ([]response.UserCouponTemplate, int64, error) {
	var list []response.UserCouponTemplate
	var total int64

	// 基础过滤：未删除 + 有效期判断（相对有效期不受 end_time 限制）
	baseQuery := c.db.Model(&model.CouponTemplate{}).
		Where("is_deleted = ?", false).
		Where("usable_days > 0 OR end_time > ?", time.Now())

	// 统计总数
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// held_count 子查询：当前用户在该模板下未过期（含未用+已用）的券数，
	// 与 CountUserCoupon / ReceiveCoupon 的限领口径完全一致
	heldCountSub := c.db.Model(&model.UserCoupon{}).
		Select("COUNT(*)").
		Where("template_id = coupon_template.template_id AND user_id = ? AND status != ?", userId, "expired")

	// 主查询：SELECT 字段 + remaining_count SQL 表达式 + held_count 子查询 + 分页
	err := baseQuery.
		Select("template_id, coupon_name, coupon_type, threshold_amount, discount_amount, "+
			"per_user_limit, usable_days, start_time, end_time, "+
			"total_count - received_count AS remaining_count, "+
			"(?) AS held_count", heldCountSub).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ============================================================
// 券状态迁移（核销 / 归还）—— 全部使用条件 UPDATE + 影响行数（乐观锁）
// ============================================================

// UseCoupon 核销用户券（下单事务内调用）
// 接收值：userCouponId - 用户券ID；orderNo - 订单号（写入券上作为使用凭据）
// 返回值：rowsAffected - 影响行数，0 表示券不可用（不存在/已用/已过期），调用方应回滚事务
//
//	err - 错误信息
//
// 说明：WHERE 含 status='unused' 与 expire_at > now()——同一张券并发核销只允许一个成功；
//
//	user_id 归属校验由服务层在调用前完成（owner 不可变，无竞态）
func (c *CouponRepo) UseCoupon(userCouponId int64, orderNo string) (int64, error) {
	// 条件核销（乐观锁）：WHERE 含 status='unused' + expire_at > now()——
	// 并发下同一张券只允许一个事务核销成功（RowsAffected 判断成败）；
	// 同时写入 order_no/used_at 作为使用凭据
	baseQuery := c.db.Model(model.UserCoupon{}).
		Where("expire_at > ? AND user_coupon_id = ? AND status = ?", time.Now(), userCouponId, "unused").
		Updates(map[string]interface{}{
			"status":   "used",
			"used_at":  time.Now(),
			"order_no": orderNo,
		})
	err := baseQuery.Error
	if err != nil {
		return 0, err
	}
	return baseQuery.RowsAffected, nil
}

// RefundCoupon 归还用户券（订单取消事务内调用，幂等）
// 接收值：userCouponId - 用户券ID
// 返回值：rowsAffected - 影响行数，0 表示券非 used 状态（如已归还/已过期迁移），调用方可忽略
//
//	err - 错误信息
//
// 说明：① 清空 order_no/used_at 还原未使用状态；② WHERE status='used' 与核销条件互斥，
//
//	同一张券不可能同时被核销和归还；③ 重复归还天然失败（幂等，防 MQ 重复消费）
func (c *CouponRepo) RefundCoupon(userCouponId int64) (int64, error) {
	// 条件归还（乐观锁）：WHERE status='used' 与核销条件（status='unused'）互斥，
	// 同一张券不可能同时被核销和归还；清空 order_no/used_at 还原"未使用"状态；
	// 重复归还时条件不匹配（RowsAffected=0），天然幂等——防 MQ 重复消费
	baseQuery := c.db.Model(model.UserCoupon{}).
		Where("user_coupon_id = ? AND status = ?", userCouponId, "used").
		Updates(map[string]interface{}{
			"status":   "unused",
			"used_at":  nil,
			"order_no": nil,
		})
	err := baseQuery.Error
	if err != nil {
		return 0, err
	}
	return baseQuery.RowsAffected, nil
}
