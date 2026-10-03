package task

import (
	"context"
	"strconv"
	"time"

	"demo-shop/services/product/internal/infra/lock"
	"demo-shop/services/product/internal/repository"
	"demo-shop/services/product/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 锁名与单体一致:搬迁前后不能出现两把不同的锁同时跑
const stockReconcileLockName = "stock:reconcile"

// 闸门键 TTL。与单体一致:闸门健本就带短 TTL,DB 是权威、Redis 只是加速。
const stockGateTTLSeconds = 30

// StockReconcileService 库存闸门对账任务。
//
// 从单体 src/task/stock_reconcile.go 平移(DS-A-26 §3)。
// 以 DB(sys_product_sku)为权威,周期性把 Redis 库存闸门计数 sku:stock:*
// 收敛到与 DB 的可用库存一致。
type StockReconcileService struct {
	repo  *repository.ProductRepo
	store *redis.Redis
	locks *lock.TaskLockManager
}

// NewStockReconcileService 构造闸门对账任务。
// store 为 nil 表示 Redis 未配置 —— 此时闸门本就没启用,Start 会直接返回。
func NewStockReconcileService(repo *repository.ProductRepo, store *redis.Redis) *StockReconcileService {
	return &StockReconcileService{
		repo:  repo,
		store: store,
		locks: lock.NewTaskLockManager(store),
	}
}

// AsService 包装成可被 go-zero service group 托管的服务(支持优雅退出)
func (s *StockReconcileService) AsService(interval time.Duration) service.Service {
	return newTaskService("stock-reconcile", func(ctx context.Context) {
		_ = s.Start(ctx, interval)
	})
}

// Start 周期对账(阻塞),由 service group 托管以便优雅退出。
func (s *StockReconcileService) Start(ctx context.Context, interval time.Duration) error {
	if s.store == nil {
		logx.Info("Redis 未配置,库存闸门对账不启动(闸门本就未启用)")
		return nil
	}
	logx.Infof("启动库存闸门对账(sku:stock:*): 周期 %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logx.Info("库存闸门对账任务退出")
			return nil
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

// ReconcileOnce 立即执行一轮(手动触发/测试入口)
func (s *StockReconcileService) ReconcileOnce(ctx context.Context) {
	s.run(ctx)
}

// run 执行一轮对账(抢跨实例锁后收敛库存闸门,拿不到锁则跳过本轮)
func (s *StockReconcileService) run(ctx context.Context) {
	// 跨实例互斥;Redis 不可用降级为进程内互斥
	release, ok := s.locks.TryLock(stockReconcileLockName, 5*time.Minute)
	if !ok {
		return
	}
	defer release()

	s.reconcileSkus(ctx)
}

// reconcileSkus 遍历在用 SKU,把库存闸门键收敛到 DB 口径:
//   - 键缺失 → SETNX 回填
//   - 键存在但与 DB 有偏差 → SET 覆盖对齐
//
// 该键同时被商品详情缓存使用,双方写的都是"DB 当前可售量",互相覆盖无害。
func (s *StockReconcileService) reconcileSkus(ctx context.Context) {
	skus, err := s.repo.GetAllActiveSkuStock()
	if err != nil {
		logx.Errorf("库存闸门对账: 查询 SKU 失败: %v", err)
		return
	}
	for _, sku := range skus {
		key := utils.StockGateKey(sku.SkuId)
		val, err := s.store.GetCtx(ctx, key)
		if err != nil {
			// 键不存在(或读失败):SETNX 回填。
			// 用 SetnxExCtx 而非 SetexCtx —— 不能覆盖并发写入:
			// 对账是"补缺"不是"夺取",下轮会再对齐。
			// 与单体 FillGateCounter 同语义(那个也只是 SETNX)。
			if _, err := s.store.SetnxExCtx(ctx, key, strconv.FormatInt(sku.Stock, 10), stockGateTTLSeconds); err != nil {
				logx.Errorf("库存闸门对账: 回填失败 skuId=%d err=%v", sku.SkuId, err)
			}
			continue
		}
		if cur, err := strconv.ParseInt(val, 10, 64); err != nil || cur != sku.Stock {
			if err := s.store.SetexCtx(ctx, key, strconv.FormatInt(sku.Stock, 10), stockGateTTLSeconds); err != nil {
				logx.Errorf("库存闸门对账: 对齐失败 skuId=%d err=%v", sku.SkuId, err)
			}
		}
	}
	logx.Infof("库存闸门对账: %d 个 SKU 收敛完成", len(skus))
}
