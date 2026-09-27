package service

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
)

type CouponService struct {
	CouponRepo *repository.CouponRepo
	db         *gorm.DB
	cache      *cache.RedisService
}

// NewCouponService 新建优惠券模块服务层实例
// 接收值：无接收值，使用全局数据库连接
// 返回值：*CouponService - 优惠券模块服务层实例指针
func NewCouponService() *CouponService {
	return &CouponService{
		CouponRepo: repository.NewCouponRepo(),
		db:         db.DB,
		cache:      infra.GetGateCache(),
	}
}

func NewCouponServiceWithCache(c *cache.RedisService) *CouponService {
	svc := NewCouponService()
	svc.cache = c
	return svc
}

// CreateCoupon 创建优惠券模板（接口1）
// 路由映射：POST /api/v1/admin/platform/coupons
// 所需权限：platform:coupon:create
// 接收值：req - 创建参数（名称/类型/门槛/优惠金额/总量/限领数/有效期）
// 返回值：*response.CreateCouponResp - 新建模板ID
//
//	error - 参数非法返回 ErrCouponParamInvalid(11007)；
//	        有效期两种模式均未配置返回 ErrCouponValidityInvalid(11008)
//
// 说明：有效期二选一——usable_days>0 表示领取后 N 天有效（相对）；
//
//	usable_days=0 时要求 start_time/end_time 成对配置（固定）
func (c *CouponService) CreateCoupon(req requset.CreateCouponReq) (*response.CreateCouponResp, error) {
	// ---- 参数校验链：任一不合法直接返回 11007 ----
	// 类型必须为两种枚举之一（与数据库 CHECK 约束 ck_coupon_type 对应）
	if req.CouponType != "full_reduction" && req.CouponType != "direct_discount" {
		return nil, model.ErrCouponParamInvalid
	}
	// 优惠力度必须 > 0（满减为金额、直减为折扣率，均不允许 0/负数）
	if req.DiscountAmount <= 0 {
		return nil, model.ErrCouponParamInvalid
	}
	// 使用门槛 >= 0（0 表示无门槛券）
	if req.ThresholdAmount < 0 {
		return nil, model.ErrCouponParamInvalid
	}
	// 发放总量必须 > 0
	if req.TotalCount <= 0 {
		return nil, model.ErrCouponParamInvalid
	}
	// 每人限领不能超过发放总量（否则限领失去意义）
	if req.PerUserLimit > req.TotalCount {
		return nil, model.ErrCouponParamInvalid
	}
	// ---- 有效期校验：两种模式必须二选一 ----
	// 相对有效期（usable_days>0）与固定有效期（start/end）均未配置 → 11008
	if req.UsableDays == 0 && (req.StartTime.IsZero() || req.EndTime.IsZero()) {
		return nil, model.ErrCouponValidityInvalid
	}

	// 校验通过后落库：received_count 走数据库默认值 0，后续领取时原子扣减
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

	// 返回新建模板 ID，供前端后续领取使用
	return &response.CreateCouponResp{
		TemplateId: templateId,
	}, nil
}

// GetCouponList 分页查询优惠券模板列表（接口2）
// 路由映射：GET /api/v1/admin/platform/coupons
// 所需权限：platform:coupon:view
// 接收值：req - 分页参数与筛选条件（coupon_name/coupon_type）
// 返回值：*response.GetCouponListResp - 模板列表 + 总数 + 分页信息
//
//	error - 数据库异常
func (c *CouponService) GetCouponList(req requset.GetCouponListReq) (*response.GetCouponListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
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

// UserGetCouponList 分页查询当前用户的优惠券列表（接口4）
// 路由映射：GET /api/v1/users/platform/coupons
// 鉴权：JWT（userId 从上下文获取，不支持越权查询他人券）
// 接收值：userId - 当前登录用户ID；req - 分页与状态筛选（unused/used/expired）
// 返回值：*response.UserGetCouponListResp - 用户券列表 + 总数
//
//	error - 数据库异常
func (c *CouponService) UserGetCouponList(userId int64, req requset.UserGetCouponListReq) (*response.UserGetCouponListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
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

// ReceiveCoupon 用户领取优惠券（接口3）—— 并发安全核心方法
// 路由映射：POST /api/v1/users/platform/coupons/receive/:templateId
// 鉴权：JWT（userId 从上下文获取）
// 接收值：userId - 当前登录用户ID；templateId - 优惠券模板ID
// 返回值：*response.UserReceiveCouponResp - 用户券ID + 过期时间
//
//	error - 模板不存在(11001)/已领完(11002)/已达上限(11003)
//
// 并发设计（双防线）：
//
//	① 悲观锁：SELECT ... FOR UPDATE 锁模板行，同一模板的并发领取串行排队；
//	   锁内校验限领（CountUserCoupon）与插入成为原子操作——防每人超领
//	② 乐观锁：UPDATE ... WHERE received_count < total_count 条件扣减，
//	   影响行数 0 即售罄——防总量超发
func (c *CouponService) receiveCoupon(userId, templateId int64) (*response.UserReceiveCouponResp, error) {
	var resp response.UserReceiveCouponResp

	gatePassed := false
	if c.cache != nil {
		ctx := context.Background()
		code, err := c.cache.DeductCouponReceive(ctx, templateId, userId)
		if err != nil {
			log.Printf("[WARN] 领券闸门异常,降级直走 DB: templateId=%d err=%v", templateId, err)
		} else {
			if code == cache.GateBackfill {
				if c.backfillCouponGate(ctx, templateId) {
					code, err = c.cache.DeductCouponReceive(ctx, templateId, userId)
				}
			}
			if err == nil {
				switch {
				case code > 0:
					gatePassed = true
				case code == cache.GateSoldOut:
					return nil, model.ErrCouponSoldOut
				case code == cache.GateLimitExceeded:
					return nil, model.ErrCouponLimitExceeded
				}
				// code 仍为 Backfill(回填失败),落到下方照常走 DB(降级语义)
			}
		}
	}

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
	// 闸门补偿，若DB失败说明本次领取并未发生，所以把阀门扣减的归还
	if err != nil {
		if gatePassed && c.cache != nil {
			if cErr := c.cache.CompensateCouponReceive(context.Background(), templateId, userId); cErr != nil {
				log.Printf("[WARN] 领券闸门补偿失败(等待对账收敛): templateId=%d err=%v", templateId, cErr)
			}
		}
		return nil, err
	}
	return &resp, nil
}

func (c *CouponService) backfillCouponGate(ctx context.Context, templateId int64) bool {
	tpl, err := c.CouponRepo.GetTemplateById(templateId)
	if err != nil {
		return false
	}
	remaining := tpl.TotalCount - tpl.ReceivedCount
	if remaining < 0 {
		remaining = 0
	}
	ttl := getTTL(tpl)
	c.cache.FillGateCounter(ctx, fmt.Sprintf("coupon:stock:%d", templateId), remaining, ttl)
	c.cache.FillGateCounter(ctx, fmt.Sprintf("coupon:limit:%d", templateId), tpl.PerUserLimit, ttl)
	return true
}

// gateTTL 闸门键存活时间 = 模板剩余有效期 + 1 天缓冲,下限 1 小时。
func getTTL(tpl *model.CouponTemplate) time.Duration {
	var d time.Duration
	if tpl.UsableDays > 0 {
		d = time.Duration(tpl.UsableDays)*24*time.Hour + 24*time.Hour
	} else if !tpl.EndTime.IsZero() {
		d = time.Until(tpl.EndTime) + 24*time.Hour
	}
	if d < time.Hour {
		d = time.Hour
	}
	return d
}

// GetReceiveCouponList 领券中心模板列表（用户端可见可领取的券）
// 路由映射：GET /api/v1/users/platform/coupons/templates
// 鉴权：JWT（userId 从上下文获取）
// 接收值：userId - 当前登录用户ID；req - 分页参数
// 返回值：*response.UserCouponTemplateListResp - 模板列表（含当前用户已领数/剩余数）
//
// 说明：① 复用 repo.UserGetTemplateList 独立查询（过滤已结束的固定有效期券）;
//
//	② 空列表是正常业务状态（前端展示"暂无可用券"），不返回错误
func (c *CouponService) GetReceiveCouponList(userId int64, req requset.UserGetTemplateListReq) (*response.UserCouponTemplateListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	list, total, err := c.CouponRepo.UserGetTemplateList(userId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &response.UserCouponTemplateListResp{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAvailableCouponList 结算时查询可用优惠券（接口5）—— 按实付金额升序
// 路由映射：GET /api/v1/users/platform/coupons/available?order_amount=xxx
// 鉴权：JWT（userId 从上下文获取）
// 接收值：userId - 当前登录用户ID；req - 订单总金额（未抵扣前）
// 返回值：*response.GetAvailableCouponResp - 可用券列表（按 PayAfter 升序）
//
//	error - order_amount 小于 0 返回 ErrCouponOrderAmountInvalid(11009)
//
// 业务规则：
//
//	① 数据源只含未过期且 unused 的券（repo 层过滤）
//	② 满减：PayAfter = 金额 - 优惠；直减：PayAfter = 金额 × 折扣率
//	③ 不满足门槛（threshold > order_amount）的券直接过滤，不进入列表
//	④ 升序排序后第一张即"最优券"，由前端标记展示
func (c *CouponService) GetAvailableCouponList(userId int64, req requset.GetAvailableCouponReq) (*response.GetAvailableCouponResp, error) {
	if req.OrderAmount < 0 {
		return nil, model.ErrCouponOrderAmountInvalid
	}
	coupons, err := c.CouponRepo.GetUserCouponList(userId)
	if err != nil {
		return nil, err
	}
	// 预分配容量 = 券数量，避免 append 扩容拷贝；长度 0 保证从空开始追加
	list := make([]response.GetAvailableCouponList, 0, len(coupons))
	for i := range coupons {
		coupon := &coupons[i]
		var totalAmount float64
		// 门槛判断：实付金额 >= 门槛才可用，否则视为不可用券直接跳过（不进入列表）
		if req.OrderAmount >= coupon.ThresholdAmount {
			// 按券类型计算实付：满减 = 原价 - 优惠金额；直减 = 原价 × 折扣率
			if coupon.CouponType == "full_reduction" {
				totalAmount = req.OrderAmount - coupon.DiscountAmount
			} else {
				totalAmount = req.OrderAmount * coupon.DiscountAmount
			}
		} else {
			continue
		}
		// 组装响应项：PayAfter 即该券用后的实付金额
		list = append(list, response.GetAvailableCouponList{
			UserCouponId:    coupon.UserCouponId,
			CouponName:      coupon.CouponName,
			CouponType:      coupon.CouponType,
			ThresholdAmount: coupon.ThresholdAmount,
			DiscountAmount:  coupon.DiscountAmount,
			PayAfter:        totalAmount,
		})

	}
	// 按实付金额升序排列：排在最前的即"最优券"，由前端标记展示
	sort.Slice(list, func(i, j int) bool {
		return list[i].PayAfter < list[j].PayAfter
	})

	return &response.GetAvailableCouponResp{
		List: list,
	}, nil

}
