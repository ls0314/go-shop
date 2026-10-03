package task

import (
	"context"
	"time"

	"demo-shop/services/marketing/internal/infra/gate"
	"demo-shop/services/marketing/internal/infra/lock"
	"demo-shop/services/marketing/internal/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// couponReconcileLockName 锁名与单体一致:搬迁前后不能出现两把不同的锁同时跑
const couponReconcileLockName = "coupon:reconcile"

// CouponReconcileService 券闸门对账任务。
//
// 从单体 src/task/coupon_reconcile.go 平移(DS-A-26 §3)。
// 以 DB 为权威,周期性做两件事:
//  1. 把已过期但仍标 unused 的券置为 expired —— 单体靠查询条件过滤、状态不迁移;
//     拆出后 CountUserCoupon 只看 status,不迁就会让过期券永久占用用户限领额度
//  2. 把 coupon:stock:{tid} / coupon:ucnt:{tid}:{uid} 收敛到 DB 口径,
//     并把 coupon:limit:{tid} 按模板定义回填
type CouponReconcileService struct {
	couponRepo     *repository.CouponRepo
	userCouponRepo *repository.UserCouponRepo
	gate           *gate.CouponGate
	locks          *lock.TaskLockManager
	store          *redis.Redis
}

// NewCouponReconcileService 构造对账任务
func NewCouponReconcileService(couponRepo *repository.CouponRepo, userCouponRepo *repository.UserCouponRepo, store *redis.Redis) *CouponReconcileService {
	return &CouponReconcileService{
		couponRepo:     couponRepo,
		userCouponRepo: userCouponRepo,
		gate:           gate.NewCouponGate(store),
		locks:          lock.NewTaskLockManager(store),
		store:          store,
	}
}

// AsService 包装成可被 go-zero service group 托管的服务(支持优雅退出)
func (s *CouponReconcileService) AsService(interval time.Duration) service.Service {
	return newTaskService("coupon-reconcile", func(ctx context.Context) {
		_ = s.Start(ctx, interval)
	})
}

// Start 周期对账(阻塞)
func (s *CouponReconcileService) Start(ctx context.Context, interval time.Duration) error {
	if s.store == nil {
		logx.Info("Redis 未配置,券闸门对账不启动(闸门本就未启用)")
		return nil
	}
	logx.Infof("启动券闸门对账(coupon:*): 周期 %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logx.Info("券闸门对账任务退出")
			return nil
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

// ReconcileOnce 立即执行一轮(手动触发/测试入口)
func (s *CouponReconcileService) ReconcileOnce(ctx context.Context) {
	s.run(ctx)
}

// run 执行一轮对账(抢跨实例锁;拿不到锁则跳过本轮)
func (s *CouponReconcileService) run(ctx context.Context) {
	release, ok := s.locks.TryLock(couponReconcileLockName, 5*time.Minute)
	if !ok {
		return
	}
	defer release()

	s.expireCoupons(ctx)
	s.reconcileGate(ctx)
}

// expireCoupons 状态迁移:unused 且已过期 → expired。
// 放在对账里而不是查询时判断,是为了与 CountUserCoupon 的口径统一(它只看 status)。
func (s *CouponReconcileService) expireCoupons(ctx context.Context) {
	rows, err := s.userCouponRepo.UpdateExpiredCoupon()
	if err != nil {
		logx.Errorf("券对账: 过期迁移失败: %v", err)
		return
	}
	if rows > 0 {
		logx.Infof("券对账: %d 张券迁移为 expired", rows)
	}
}

// reconcileGate 把闸门计数收敛到 DB 口径。
//
// 与库存对账的差异:券闸门有两个键要维护(stock 余量 + 每用户已领数),
// 且 limit 键由模板定义、不从 DB 逐行算 —— 但它也纳入对账,防止
// 有人改了 per_user_limit 而闸门还停在旧值。
func (s *CouponReconcileService) reconcileGate(ctx context.Context) {
	templates, err := s.couponRepo.GetActiveCouponList()
	if err != nil {
		logx.Errorf("券对账: 查询模板失败: %v", err)
		return
	}

	for _, tpl := range templates {
		ttl := gate.GateTTL(tpl)

		// 余量:缺失则回填,有偏差则覆盖对齐。
		// 覆盖而非 SETNX:这里是"以 DB 为准收敛",与领取路径的回填语义不同 ——
		// 回填是怕并发覆盖刚扣的进度,对账是专门来纠正偏差的。
		remaining := tpl.TotalCount - tpl.ReceivedCount
		if remaining < 0 {
			remaining = 0
		}
		stockKey := gate.StockKey(tpl.TemplateId)
		if val, err := s.store.GetCtx(ctx, stockKey); err != nil {
			s.gate.Fill(ctx, stockKey, remaining, ttl)
		} else if cur, perr := parseInt(val); perr != nil || cur != remaining {
			if err := s.store.SetexCtx(ctx, stockKey, itoa(remaining), int(ttl.Seconds())); err != nil {
				logx.Errorf("券对账: 余量对齐失败 templateId=%d err=%v", tpl.TemplateId, err)
			}
		}

		// 每人限领数:模板定义变了就跟着改(这是唯一会改 limit 键的地方)
		limitKey := gate.LimitKey(tpl.TemplateId)
		limitVal := itoa(tpl.PerUserLimit)
		if val, err := s.store.GetCtx(ctx, limitKey); err != nil {
			s.gate.Fill(ctx, limitKey, tpl.PerUserLimit, ttl)
		} else if val != limitVal {
			if err := s.store.SetexCtx(ctx, limitKey, limitVal, int(ttl.Seconds())); err != nil {
				logx.Errorf("券对账: 限领数对齐失败 templateId=%d err=%v", tpl.TemplateId, err)
			}
		}
	}

	logx.Infof("券闸门对账: %d 个模板收敛完成", len(templates))
}
