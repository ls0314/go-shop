package task

import (
	"context"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"fmt"
	"log"
	"strconv"
	"time"
)

// CouponReconcileService 券闸门对账任务
//
// 以 DB(coupon_template)为权威,周期性把 Redis 券闸门计数收敛到与 DB 一致:
// coupon:stock:* 对齐 total_count - received_count,coupon:limit:* 为每用户限领数。
// ucnt(每用户已领数)不对账。
type CouponReconcileService struct {
	couponRepo *repository.CouponRepo
	cache      *cache.RedisService
	locks      *DistributedLockManager
}

func NewCouponReconcileService(deps service.ServiceDeps) *CouponReconcileService {
	return &CouponReconcileService{
		couponRepo: repository.NewCouponRepo(deps.DB),
		cache:      deps.Cache,
		locks:      NewDistributedLockManager(deps.Cache),
	}
}

func NewCouponReconcileServiceWithCache(deps service.ServiceDeps, c *cache.RedisService) *CouponReconcileService {
	return &CouponReconcileService{
		couponRepo: repository.NewCouponRepo(deps.DB),
		cache:      c,
		locks:      NewDistributedLockManager(c),
	}
}

// ReconcileOnce 立即执行一轮对账(测试/手动触发入口;周期入口见 Start)
func (s *CouponReconcileService) ReconcileOnce() {
	s.run()
}

// Start 周期对账(放 goroutine)
func (s *CouponReconcileService) Start(interval time.Duration) {
	if s.cache == nil {
		log.Println("[INFO] Redis 未初始化,券闸门对账不启动(闸门本就未启用)")
		return
	}
	log.Printf("[INFO] 启动券闸门对账(coupon:*): 周期 %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		go s.run() // 异步:对账耗时不影响下一轮节拍
	}
}

// run 执行一轮对账(抢跨实例锁后收敛券闸门,拿不到锁则跳过本轮)
func (s *CouponReconcileService) run() {
	// 跨实例互斥(Redsync);Redis 不可用降级为进程内互斥
	release, ok := s.locks.TryLock("coupon:reconcile", 5*time.Minute)
	if !ok {
		return
	}
	defer release()

	ctx := context.Background()
	s.reconcileCoupons(ctx)
}

// reconcileCoupons 遍历在用模板,把券闸门键收敛到 DB 口径:
//   - 键缺失 → SETNX 回填 stock 与 limit
//   - 键存在但与 DB 有偏差 → SET 覆盖对齐
//   - limit 键只在缺失时补(不可变字段)
func (s *CouponReconcileService) reconcileCoupons(ctx context.Context) {
	templates, err := s.couponRepo.GetAllActiveTemplates()
	if err != nil {
		log.Printf("[WARN] 券闸门对账: 查询模板失败: %v", err)
		return
	}
	for i := range templates {
		tpl := &templates[i]
		remaining := tpl.TotalCount - tpl.ReceivedCount
		if remaining < 0 {
			remaining = 0
		}
		stockKey := fmt.Sprintf("coupon:stock:%d", tpl.TemplateId)
		limitKey := fmt.Sprintf("coupon:limit:%d", tpl.TemplateId)
		ttl := gateTTL(tpl)

		val, err := s.cache.Get(ctx, stockKey)
		if err != nil { // 含 redis.Nil(缺键) → 回填补位
			s.cache.FillGateCounter(ctx, stockKey, remaining, ttl)
			s.cache.FillGateCounter(ctx, limitKey, tpl.PerUserLimit, ttl)
			continue
		}
		if cur, err := strconv.ParseInt(val, 10, 64); err != nil || cur != remaining {
			// 已存在但偏差 → 强制对齐。这里必须用 SET(覆盖)而非 SETNX,否则收敛不了
			_ = s.cache.Set(ctx, stockKey, strconv.FormatInt(remaining, 10), ttl)
		}
	}
	log.Printf("[INFO] 券闸门对账: %d 个模板收敛完成", len(templates))
}

// gateTTL 返回券闸门键的 TTL:按 usable_days 或 end_time 计算并多留 1 天,下限 1 小时
func gateTTL(tpl *model.CouponTemplate) time.Duration {
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
