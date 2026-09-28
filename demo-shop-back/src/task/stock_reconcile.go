package task

import (
	"context"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"fmt"
	"log"
	"strconv"

	"time"
)

// StockReconcileService 库存闸门对账任务
//
// 以 DB(sys_product_sku)为权威,周期性把 Redis 库存闸门计数 sku:stock:*
// 收敛到与 DB 的可用库存一致。
type StockReconcileService struct {
	productRepo *repository.ProductRepo
	cache       *cache.RedisService
	locks       *DistributedLockManager
}

func NewStockReconcileService(deps service.ServiceDeps) *StockReconcileService {
	return &StockReconcileService{
		productRepo: repository.NewProductRepo(deps.DB),
		cache:       deps.Cache,
		locks:       NewDistributedLockManager(deps.Cache),
	}
}

// NewStockReconcileServiceWithCache 测试专用构造:显式注入 Redis,
// 使「对账收敛」成为可被测试断言的行为(C5 阶段二)
func NewStockReconcileServiceWithCache(deps service.ServiceDeps, c *cache.RedisService) *StockReconcileService {
	return &StockReconcileService{
		productRepo: repository.NewProductRepo(deps.DB),
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
		log.Println("[INFO] Redis 未初始化,库存闸门对账不启动(闸门本就未启用)")
		return
	}
	log.Printf("[INFO] 启动库存闸门对账(sku:stock:*): 周期 %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		go s.run() // 异步:对账耗时不影响下一轮节拍
	}
}

// run 执行一轮对账(抢跨实例锁后收敛库存闸门,拿不到锁则跳过本轮)
func (s *StockReconcileService) run() {
	// 跨实例互斥(Redsync);Redis 不可用降级为进程内互斥
	release, ok := s.locks.TryLock("stock:reconcile", 5*time.Minute)
	if !ok {
		return
	}
	defer release()

	ctx := context.Background()
	s.reconcileSkus(ctx)
}

// reconcileSkus 遍历在用 SKU,把库存闸门键收敛到 DB 口径:
//   - 键缺失 → SETNX 回填
//   - 键存在但与 DB 有偏差 → SET 覆盖对齐
//
// 该键同时被商品详情缓存使用,双方写的都是"DB 当前可售量",互相覆盖无害。
func (s *StockReconcileService) reconcileSkus(ctx context.Context) {
	skus, err := s.productRepo.GetAllActiveSkuStock()
	if err != nil {
		log.Printf("[WARN] 库存闸门对账: 查询 SKU 失败: %v", err)
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
	log.Printf("[INFO] 库存闸门对账: %d 个 SKU 收敛完成", len(skus))
}
