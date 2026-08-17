package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepo struct {
	db *gorm.DB
}

func NewCouponRepo() *CouponRepo {
	return &CouponRepo{
		db: db.DB,
	}
}

func (c *CouponRepo) WithTx(tx *gorm.DB) *CouponRepo {
	return &CouponRepo{
		db: tx,
	}
}

func (c *CouponRepo) CreateCoupon(coupon model.CouponTemplate) (int64, error) {
	err := c.db.Create(&coupon).Error
	return coupon.TemplateId, err
}

func (c *CouponRepo) CreateUserCoupon(userCoupon model.UserCoupon) (int64, error) {
	err := c.db.Create(&userCoupon).Error
	return userCoupon.UserCouponId, err
}

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

// CountUserCoupon 统计用户在指定模板下占用名额的券数(排除已过期)
func (c *CouponRepo) CountUserCoupon(userId, templateId int64) (int64, error) {
	var cnt int64
	err := c.db.Model(&model.UserCoupon{}).
		Where("user_id = ? AND template_id = ? AND status != ?", userId, templateId, "expired").
		Count(&cnt).Error
	return cnt, err
}

func (c *CouponRepo) GetUserCouponList(userId int64) ([]response.UserGetCouponList, error) {
	var coupons []response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
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

func (c *CouponRepo) GetUserCoupon(userCouponId int64) (*response.UserGetCouponList, error) {
	var coupon response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
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

func (c *CouponRepo) GetCouponList(req *requset.GetCouponListReq) ([]response.GetCouponList, int64, error) {
	var coupons []response.GetCouponList

	baseQuery := c.db.Model(&model.CouponTemplate{}).
		Where("is_deleted = ?", false)

	if req.CouponName != "" {
		baseQuery = baseQuery.Where("coupon_name = ?", req.CouponName)
	}
	if req.CouponType != "" {
		baseQuery = baseQuery.Where("coupon_type = ?", req.CouponType)
	}

	var total int64
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize
	err = baseQuery.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&coupons).Error
	if err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

func (c *CouponRepo) UserGetCouponList(userId int64, req *requset.UserGetCouponListReq) ([]response.UserGetCouponList, int64, error) {
	var coupons []response.UserGetCouponList
	templateTable := model.CouponTemplate{}.TableName()
	userCouponTable := model.UserCoupon{}.TableName()
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

// LockTemplateByID 事务内以 FOR UPDATE 锁定模板行,并读取锁内最新数据
// 必须由事务内的 CouponRepo 调用(WithTx 之后),否则语句结束锁即释放
func (c *CouponRepo) LockTemplateByID(templateId int64) (*model.CouponTemplate, error) {
	var tpl model.CouponTemplate
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
func (c *CouponRepo) UseCoupon(userCouponId int64, orderNo string) (int64, error) {
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
