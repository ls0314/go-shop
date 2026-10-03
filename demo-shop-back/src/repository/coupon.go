package repository

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// 优惠券模块数据层
// ============================================================
//
// ⚠️ **本文件是过渡态,归属 C4**。
//
// 券表(coupon_template / user_coupon)的所有权已于 C3 迁至 marketing-service
// (独立库 marketing_db)。券域的**读接口与领券**已全部改经 RPC
// (见 infra/couponclient + service/coupon_service.go),此处只余
// **下单链路仍在用的四个方法**:
//
//	GetUserCoupon           下单前取券(校验归属 + 拿门槛/优惠)
//	UseCoupon               下单核销(写入 order_no)
//	GetUserCouponByOrderNo  取消前按订单号反查券
//	RefundCoupon            取消时归还券
//
// 它们之所以还在:这四处调用位于**下单/取消的本地事务内**,与订单创建同生共死。
// 改成 RPC 就变跨服务两步操作,需要 Saga 补偿(核销成功而本地事务回滚 →
// 必须 ReturnCoupon)—— 那是 C4 的核心工作(见待办 §9.3)。
//
// 在 C4 完成之前,**这段代码读写的是 demo_shop 里的旧券表**,与
// marketing-service 的 marketing_db 是两个库。当前两库券表都是 0 行,
// 尚无数据损坏;但一旦经 RPC 建了券模板就会分裂 —— 故券域在 C4 前
// 不可做端到端验证。
type CouponRepo struct {
	db *gorm.DB
}

// NewCouponRepo 新建优惠券模块数据层实例
// 接收值：conn - 数据库连接
// 返回值：*CouponRepo - 优惠券模块数据层实例指针
func NewCouponRepo(conn *gorm.DB) *CouponRepo {
	return &CouponRepo{db: conn}
}

// WithTx 优惠券表事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*CouponRepo - 绑定事务的优惠券模块数据层指针
// 说明：事务内所有方法必须使用本方法返回的实例，保证读写同属一个事务
func (c *CouponRepo) WithTx(tx *gorm.DB) *CouponRepo {
	return &CouponRepo{db: tx}
}

// ============================================================
// 下单链路在用(C4 前不可删)
// ============================================================

// GetUserCoupon 事务内按ID取券(下单核销前的校验)
// 接收值：userCouponId - 用户券ID
// 返回值：*response.UserGetCouponList - 券信息（含归属 UserId 与门槛/优惠）
//
//	err - 无匹配记录返回 model.ErrCouponNotExistOrUsed
func (c *CouponRepo) GetUserCoupon(userCouponId int64) (*response.UserGetCouponList, error) {
	var coupon response.UserGetCouponList
	// 联模板取券规则;同时过滤掉已过期/非 unused 的券 —— 与 UseCoupon 的条件一致
	err := c.db.Model(&model.UserCoupon{}).
		Select("user_coupon.user_coupon_id, user_coupon.user_id, user_coupon.status, "+
			"user_coupon.order_no, user_coupon.expire_at, "+
			"coupon_template.coupon_name, coupon_template.coupon_type, "+
			"coupon_template.threshold_amount, coupon_template.discount_amount").
		Joins("LEFT JOIN coupon_template ON coupon_template.template_id = user_coupon.template_id").
		Where("user_coupon.user_coupon_id = ? AND user_coupon.status = ? AND user_coupon.expire_at > ?",
			userCouponId, model.CouponUnused, time.Now()).
		First(&coupon).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrCouponNotExistOrUsed
		}
		return nil, err
	}
	return &coupon, nil
}

// GetUserCouponByOrderNo 按订单号反查券（取消订单归还用）
// 接收值：orderNo - 订单号
// 返回值：*model.UserCoupon - 用户券实体
//
//	err - 无匹配记录返回 model.ErrCouponNotExistOrUsed
func (c *CouponRepo) GetUserCouponByOrderNo(orderNo string) (*model.UserCoupon, error) {
	var coupon model.UserCoupon
	err := c.db.Where("order_no = ?", orderNo).First(&coupon).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrCouponNotExistOrUsed
		}
		return nil, err
	}
	return &coupon, nil
}

// UseCoupon 核销用户券（订单创建事务内调用，条件更新防重复核销）
// 接收值：userCouponId - 用户券ID；orderNo - 关联订单号
// 返回值：rowsAffected - 影响行数，0 表示券不可用（已被核销/过期）
//
//	err - 错误信息
//
// 说明：条件 WHERE status='unused' 与 expire_at > now() 由数据库原子判定，
//
//	并发下同一张券只有一个事务能成功（乐观锁）；不用 SELECT ... FOR UPDATE
//	是因为核销是单行终态变更，条件更新足够且更轻
func (c *CouponRepo) UseCoupon(userCouponId int64, orderNo string) (int64, error) {
	baseQuery := c.db.Model(model.UserCoupon{}).
		Where("expire_at > ? AND user_coupon_id = ? AND status = ?",
			time.Now(), userCouponId, "unused").
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
