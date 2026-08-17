package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"sort"
	"time"

	"gorm.io/gorm"
)

type CouponService struct {
	CouponRepo *repository.CouponRepo
	db         *gorm.DB
}

func NewCouponService() *CouponService {
	return &CouponService{
		CouponRepo: repository.NewCouponRepo(),
		db:         db.DB,
	}
}

func (c *CouponService) CreateCoupon(req requset.CreateCouponReq) (*response.CreateCouponResp, error) {
	if req.CouponType != "full_reduction" && req.CouponType != "direct_discount" {
		return nil, model.ErrCouponParamInvalid
	}
	if req.DiscountAmount <= 0 {
		return nil, model.ErrCouponParamInvalid
	}
	if req.ThresholdAmount < 0 {
		return nil, model.ErrCouponParamInvalid
	}
	if req.TotalCount <= 0 {
		return nil, model.ErrCouponParamInvalid
	}
	if req.PerUserLimit > req.TotalCount {
		return nil, model.ErrCouponParamInvalid
	}
	if req.UsableDays == 0 && (req.StartTime.IsZero() || req.EndTime.IsZero()) {
		return nil, model.ErrCouponValidityInvalid
	}

	templateId, err := c.CouponRepo.CreateCoupon(model.CouponTemplate{
		CouponName:      req.CouponName,
		CouponType:      req.CouponType,
		ThresholdAmount: req.ThresholdAmount,
		DiscountAmount:  req.DiscountAmount,
		TotalCount:      req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		UsableDays:      req.UsableDays,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
	})
	if err != nil {
		return nil, err
	}

	return &response.CreateCouponResp{
		TemplateId: templateId,
	}, nil
}

func (c *CouponService) GetCouponList(req requset.GetCouponListReq) (*response.GetCouponListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	couponList, total, err := c.CouponRepo.GetCouponList(&req)
	if err != nil {
		return nil, err
	}
	return &response.GetCouponListResp{
		List:     couponList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (c *CouponService) UserGetCouponList(userId int64, req requset.UserGetCouponListReq) (*response.UserGetCouponListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	couponList, total, err := c.CouponRepo.UserGetCouponList(userId, &req)
	if err != nil {
		return nil, err
	}
	return &response.UserGetCouponListResp{
		List:     couponList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (c *CouponService) ReceiveCoupon(userId, templateId int64) (*response.UserReceiveCouponResp, error) {
	var resp response.UserReceiveCouponResp

	err := c.db.Transaction(func(tx *gorm.DB) error {
		couponTx := c.CouponRepo.WithTx(tx)

		// 锁定模板行 —— 从这里开始,同一模板的并发领取在此排队
		tpl, err := couponTx.LockTemplateByID(templateId)
		if err != nil {
			return err // 模板不存在时 err 即 ErrCouponTemplateNotExist
		}

		// 锁内校验(此刻其他领取事务都在等锁,读到的必是最新数据)
		if tpl.IsDeleted {
			return model.ErrCouponTemplateNotExist
		}
		if tpl.ReceivedCount >= tpl.TotalCount {
			return model.ErrCouponSoldOut
		}
		held, err := couponTx.CountUserCoupon(userId, templateId)
		if err != nil {
			return err
		}
		if held >= tpl.PerUserLimit {
			return model.ErrCouponLimitExceeded
		}

		// 原子条件扣减(行锁 + 条件更新双保险)
		rowsAffected, err := couponTx.ReceiveCoupon(templateId, tpl.TotalCount)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return model.ErrCouponSoldOut
		}

		// 计算过期时间并插入用户券(数据全部来自锁内读到的 tpl)
		var expireAt time.Time
		if tpl.UsableDays > 0 {
			expireAt = time.Now().AddDate(0, 0, int(tpl.UsableDays))
		} else {
			expireAt = tpl.EndTime
		}
		userCouponId, err := couponTx.CreateUserCoupon(model.UserCoupon{
			TemplateId: templateId,
			UserId:     userId,
			ExpireAt:   expireAt,
			Status:     model.CouponUnused,
		})
		if err != nil {
			return err
		}

		resp = response.UserReceiveCouponResp{
			UserCouponId: userCouponId,
			ExpireTime:   expireAt,
		}
		return nil //commit:锁释放,下一个排队的领取事务开始执行
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *CouponService) GetAvailableCouponList(userId int64, req requset.GetAvailableCouponReq) (*response.GetAvailableCouponResp, error) {
	if req.OrderAmount < 0 {
		return nil, model.ErrCouponOrderAmountInvalid
	}
	coupons, err := c.CouponRepo.GetUserCouponList(userId)
	if err != nil {
		return nil, err
	}
	list := make([]response.GetAvailableCouponList, 0, len(coupons))
	for i := range coupons {
		coupon := &coupons[i]
		var totalAmount float64
		if req.OrderAmount >= coupon.ThresholdAmount {
			if coupon.CouponType == "full_reduction" {
				totalAmount = req.OrderAmount - coupon.DiscountAmount
			} else {
				totalAmount = req.OrderAmount * coupon.DiscountAmount
			}
		} else {
			continue
		}
		list = append(list, response.GetAvailableCouponList{
			UserCouponId:    coupon.UserCouponId,
			CouponName:      coupon.CouponName,
			CouponType:      coupon.CouponType,
			ThresholdAmount: coupon.ThresholdAmount,
			DiscountAmount:  coupon.DiscountAmount,
			PayAfter:        totalAmount,
		})

	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].PayAfter < list[j].PayAfter
	})

	return &response.GetAvailableCouponResp{
		List: list,
	}, nil

}
