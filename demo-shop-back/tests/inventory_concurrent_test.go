package tests

import (
	"testing"

	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
)

// I1 防超卖:stock=10,40 并发各锁 1 件(独立订单) → 恰好 10 成功 / 30 不足
func TestLockStock_Concurrent_NoOversell(t *testing.T) {
	const (
		stock      = int64(10)
		goroutines = 40
		qty        = int64(1)
	)
	skuID, _ := mustCreateSkuWithStock(t, stock)
	svc := service.NewInventoryService(testDeps())

	stats, _ := runConcurrent(t, goroutines, func(i int) (int64, error) {
		return 0, svc.LockStock(skuID, qty, int64(700000+i)) // order_id 无外键,可编
	})

	if stats["success"] != int(stock) || stats["not_enough"] != goroutines-int(stock) || stats["other"] != 0 {
		t.Fatalf("I1 断言失败: %v (期望 success=%d not_enough=%d other=0)", stats, stock, goroutines-int(stock))
	}
	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 0 {
		t.Fatalf("stock 应为 0, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != stock {
		t.Fatalf("lock_stock 应为 %d, 实际 %d", stock, got)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM sys_product_stock_log WHERE sku_id = ? AND change_type = ?`, skuID, model.StockOrderLock); got != stock {
		t.Fatalf("order_lock 流水应为 %d 条, 实际 %d", stock, got)
	}
}

// I2 操作幂等:同一订单重复 LockStock → 第二次走流水去重返回 nil,库存只挪一次
func TestLockStock_Idempotent_SameOrder(t *testing.T) {
	qty := int64(3)
	orderID := nextOrderID() // ← 跨运行唯一,不再命中历史流水(替换原 const)
	skuID, _ := mustCreateSkuWithStock(t, 10)
	svc := service.NewInventoryService(testDeps())

	if err := svc.LockStock(skuID, qty, orderID); err != nil {
		t.Fatalf("首次锁定失败: %v", err)
	}
	if err := svc.LockStock(skuID, qty, orderID); err != nil {
		t.Fatalf("重复锁定应幂等返回 nil, 实际: %v", err)
	}
	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 10-qty {
		t.Fatalf("stock 应为 %d, 实际 %d", 10-qty, got)
	}
	if got := queryInt64(t,
		`SELECT COUNT(*) FROM sys_product_stock_log WHERE sku_id = ? AND order_id = ? AND change_type = ?`,
		skuID, orderID, model.StockOrderLock); got != 1 {
		t.Fatalf("order_lock 流水应为 1 条, 实际 %d", got)
	}
}

// I3 释放库存幂等:重复 ReleaseStock → 库存只回补一次,release 流水恰 1 条
func TestReleaseStock_Idempotent(t *testing.T) {
	qty := int64(3)
	orderID := nextOrderID() // ← 跨运行唯一,不再命中历史流水(替换原 const)
	skuID, _ := mustCreateSkuWithStock(t, 10)
	svc := service.NewInventoryService(testDeps())

	if err := svc.LockStock(skuID, qty, orderID); err != nil {
		t.Fatalf("前置锁定失败: %v", err)
	}

	if err := svc.ReleaseStock(skuID, qty, orderID); err != nil {
		t.Fatalf("首次释放失败: %v", err)
	}
	if err := svc.ReleaseStock(skuID, qty, orderID); err != nil {
		t.Fatalf("重复释放应幂等返回 nil, 实际: %v", err)
	}
	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 10 {
		t.Fatalf("释放后 stock 应回 10, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 0 {
		t.Fatalf("释放后 lock_stock 应为 0, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM sys_product_stock_log WHERE sku_id = ? AND order_id = ? AND change_type = ?`,
		skuID, orderID, model.StockOrderLock); got != 1 {
		t.Fatalf("order_release 流水应为 1 条, 实际 %d", got)
	}
}

// I4 多 SKU 订单:同一订单锁两个不同 SKU → 两个都必须真实锁定
// (修复前 CheckOrderLogExists 只按 order_id 判重,第二个 SKU 会被幂等跳过——生产 bug 回归测试)
func TestLockStock_MultiSku_SameOrder(t *testing.T) {
	orderID := nextOrderID()
	sku1, _ := mustCreateSkuWithStock(t, 5)
	sku2, _ := mustCreateSkuWithStock(t, 5)
	svc := service.NewInventoryService(testDeps())

	if err := svc.LockStock(sku1, 2, orderID); err != nil {
		t.Fatalf("锁 SKU1 失败: %v", err)
	}
	if err := svc.LockStock(sku2, 1, orderID); err != nil {
		t.Fatalf("锁 SKU2 失败: %v", err) // 修复前:这里被幂等跳过,静默成功
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, sku1); got != 2 {
		t.Fatalf("SKU1 lock_stock 应为 2, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, sku2); got != 1 {
		t.Fatalf("SKU2 lock_stock 应为 1, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM sys_product_stock_log WHERE order_id = ? AND change_type = ?`, orderID, model.StockOrderLock); got != 2 {
		t.Fatalf("同订单应有 2 条 order_lock 流水, 实际 %d", got)
	}
}

// I5 并发释放幂等:同订单同 SKU 双 goroutine 并发 ReleaseStock
// → 两次都返回 nil(一次真实释放 + 一次唯一索引幂等命中),库存只回补一次,release 流水恰 1 条
func TestReleaseStock_Concurrent_SameOrder(t *testing.T) {
	const (
		stock      = int64(10)
		goroutines = 2
		qty        = int64(1)
	)
	orderID := nextOrderID()
	skuID, _ := mustCreateSkuWithStock(t, stock)
	svc := service.NewInventoryService(testDeps())

	if err := svc.LockStock(skuID, qty, orderID); err != nil {
		t.Fatalf("前置锁库存失败：%v", err)
	}

	stats, _ := runConcurrent(t, goroutines, func(i int) (int64, error) {
		return 0, svc.ReleaseStock(skuID, qty, orderID) // order_id 无外键,可编
	})

	if stats["success"] != goroutines || stats["other"] != 0 {
		t.Fatalf("I5 断言失败: %v (期望 success=%d other=0)", stats, goroutines)
	}

	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != stock {
		t.Fatalf("stock 应回%d(仅回补一次）， 实际上%d)", got, stock)
	}

	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 0 {
		t.Fatalf("lock_stock 应为 0, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM sys_product_stock_log WHERE sku_id = ? AND order_id = ? AND change_type = ?`,
		skuID, orderID, model.StockOrderRelease); got != 1 {
		t.Fatalf("order_release 流水应为 1 条, 实际 %d", got)
	}
}

// I6 多 SKU 并发释放回归:同订单 3 个 SKU 并发各释放一次
// → 三元组唯一索引不得跨 SKU 误判重(与 I4 同类的粒度缺陷),3 条流水齐全、各自库存正确回补
func TestReleaseStock_MultiSku_SameOrder_Concurrent(t *testing.T) {
	const (
		stock = int64(5)
		qty   = int64(1)
	)
	orderID := nextOrderID()
	sku1, _ := mustCreateSkuWithStock(t, stock)
	sku2, _ := mustCreateSkuWithStock(t, stock)
	sku3, _ := mustCreateSkuWithStock(t, stock)
	skuIDs := []int64{sku1, sku2, sku3}
	svc := service.NewInventoryService(testDeps())

	for _, s := range skuIDs {
		if err := svc.LockStock(s, qty, orderID); err != nil {
			t.Fatalf("前置锁定 SKU%d 失败: %v", s, err)
		}
	}

	stats, _ := runConcurrent(t, len(skuIDs), func(i int) (int64, error) {
		return 0, svc.ReleaseStock(skuIDs[i], qty, orderID)
	})

	if stats["success"] != len(skuIDs) || stats["other"] != 0 {
		t.Fatalf("I6 断言失败: %v (期望 success=3 other=0)", stats)
	}
	for _, s := range skuIDs {
		if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, s); got != stock {
			t.Fatalf("SKU%d stock 应回 %d, 实际 %d", s, stock, got)
		}
		if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, s); got != 0 {
			t.Fatalf("SKU%d lock_stock 应为 0, 实际 %d", s, got)
		}
	}
	if got := queryInt64(t, `SELECT COUNT(*) FROM sys_product_stock_log WHERE order_id = ? AND change_type = ?`,
		orderID, model.StockOrderRelease); got != 3 {
		t.Fatalf("同订单应有 3 条 order_release 流水, 实际 %d", got)
	}
}
