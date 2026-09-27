package task

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"fmt"
	"log"
	"strconv"

	"time"
)

// StockReconcileService 闸门对账任务(DS-A-19)
// 职责:以 DB 为唯一权威,周期性收敛 Redis 闸门计数。
// 铁律:单向 DB → Redis,绝不反向 —— 闸门只是挡量的估算器,
// 没有资格改账本;反向写会把"估算"固化成"事实"
type StockReconcileService struct {
	couponRepo  *repository.CouponRepo
	productRepo *repository.ProductRepo
	cache       *cache.RedisService
	locks       *DistributedLockManager
}

func NewStockReconcileService() *StockReconcileService {
	return &StockReconcileService{
		couponRepo:  repository.NewCouponRepo(db.DB),
		productRepo: repository.NewProductRepo(db.DB),
		cache:       infra.GetCache(),
		locks:       NewDistributedLockManager(infra.GetCache()),
	}
}

// NewStockReconcileServiceWithCache 测试专用构造:显式注入 Redis,
// 使「对账收敛」成为可被测试断言的行为(C5 阶段二)
func NewStockReconcileServiceWithCache(c *cache.RedisService) *StockReconcileService {
	return &StockReconcileService{
		couponRepo:  repository.NewCouponRepo(db.DB),
		productRepo: repository.NewProductRepo(db.DB),
		cache:       c,
		locks:       NewDistributedLockManager(c),
	}
}

// ReconcileOnce 立即执行一轮对账(测试/手动触发入口;周期入口见 Start)
func (s *StockReconcileService) ReconcileOnce() {
	s.run()
}

// Start 周期对账(放 goroutine)
func (s *StockReconcileService) Start(interval time.Duration) {
	if s.cache == nil {
		log.Println("[INFO] Redis 未初始化,闸门对账不启动(闸门本就未启用)")
		return
	}
	log.Printf("[INFO] 启动Redis闸门对账: 周期 %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		go s.run() // 异步:对账耗时不影响下一轮节拍
	}
}

func (s *StockReconcileService) run() {
	// 跨实例互斥(Redsync);Redis 不可用降级为进程内互斥
	release, ok := s.locks.TryLock("stock:reconcile", 5*time.Minute)
	if !ok {
		return
	}
	defer release()

	ctx := context.Background()
	s.reconcileCoupons(ctx)
	s.reconcileSkus(ctx)
}

// reconcileCoupons 券闸门收敛:
//   - 键缺失 → SETNX 回填(键被淘汰/重启后的静默补位)
//   - 键存在但与 DB 有偏差 → SET 强制对齐(必须覆盖才能收敛;SET 与闸门 DECR 并发时
//     可能覆盖掉个别预扣 → 闸门短暂偏松 → DB 双防线兜底,偏差方向安全)
//   - limit 键只在缺失时补(不可变字段,SETNX 一次)
//   - ucnt 不对账:每用户键数量不可控,且偏差只会"偏严"(多拒),无资损风险
func (s *StockReconcileService) reconcileCoupons(ctx context.Context) {
	templates, err := s.couponRepo.GetAllActiveTemplates()
	if err != nil {
		log.Printf("[WARN] 闸门对账: 查询模板失败: %v", err)
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
	log.Println("[INFO] 闸门对账: 券模板收敛完成")
}

// reconcileSkus SKU 闸门收敛(与券同构)。
// 注意 sku:stock 键同时被详情缓存与闸门使用,双方写的都是"DB 当前可售量",
// 语义一致,互相覆盖无害 —— 这正是当初选择复用该键的原因
func (s *StockReconcileService) reconcileSkus(ctx context.Context) {
	skus, err := s.productRepo.GetAllActiveSkuStock()
	if err != nil {
		log.Printf("[WARN] 闸门对账: 查询 SKU 失败: %v", err)
		return
	}
	for _, sku := range skus {
		key := fmt.Sprintf("sku:stock:%d", sku.SkuId)
		val, err := s.cache.Get(ctx, key)
		if err != nil {
			s.cache.FillGateCounter(ctx, key, sku.Stock, 30*time.Second)
			continue
		}
		if cur, err := strconv.ParseInt(val, 10, 64); err != nil || cur != sku.Stock {
			_ = s.cache.Set(ctx, key, strconv.FormatInt(sku.Stock, 10), 30*time.Second)
		}
	}
	log.Println("[INFO] 闸门对账: SKU 收敛完成")
}

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
