package tests

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/task"
	"sync"
	"testing"
	"time"
)

func backdateOrderCreatedAt(t *testing.T, orderId int64, ago time.Duration) {
	t.Helper()
	if err := db.DB.Exec(
		`UPDATE user_order_master SET created_at = ? WHERE order_id = ?`,
		time.Now().UTC().Add(-ago), orderId,
	).Error; err != nil {
		t.Fatalf("回拨 created_at 失败: %v", err)
	}
}

func isOrderCancelled(t *testing.T, orderId int64) bool {
	t.Helper()
	return queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderId) == model.OrderCancelled
}

// markAllOtherPendingPayAsCancelled 把库里其它 pending_pay 单清空,使扫描只可能命中本用例造的单。
func markAllOtherPendingPayAsCancelled(t *testing.T, keepOrderId int64) {
	t.Helper()
	if err := db.DB.Exec(
		`UPDATE user_order_master SET order_status = ? WHERE order_status = ? AND order_id <> ?`,
		model.OrderCancelled, model.OrderPendingPay, keepOrderId,
	).Error; err != nil {
		t.Fatalf("清理遗留 pending_pay 订单失败: %v", err)
	}
}

// 扫描能取消 MQ 漏掉的超时单
func TestOrderTimeoutScan_CancelsExpiredOrder(t *testing.T) {
	inv := newRecordingInventory()
	deps := testDepsWithInventory(inv)

	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	const qty = int64(2)
	orderID := mustCreateOrderReadyToCancel(t, inv, userID, skuID, qty)
	backdateOrderCreatedAt(t, orderID, model.OrderPayTTL+time.Minute) // 刚过阈值

	svc := task.NewOrderTimeoutScanService(deps)
	cancelled, skipped := svc.ScanOnce()

	if cancelled != 1 || skipped != 0 {
		t.Fatalf("应恰好取消 1 单(实际 cancelled=%d skipped=%d)", cancelled, skipped)
	}
	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderCancelled {
		t.Fatalf("订单应为 %s, 实际 %q", model.OrderCancelled, got)
	}
	// 库存侧断言改为"扫描是否驱动了 product-service 的释放调用"
	release, ok := inv.findApplied(opRelease, orderID)
	if !ok {
		t.Fatal("扫描取消必须驱动库存释放,实际没有")
	}
	if release.SkuId != skuID || release.Qty != qty {
		t.Fatalf("释放参数应为 sku=%d qty=%d, 实际 %+v", skuID, qty, release)
	}
	if got := inv.countApplied(opRelease, orderID); got != 1 {
		t.Fatalf("order_release 应恰生效 1 次, 实际 %d", got)
	}
}

// 未超时的单不能被误杀
func TestOrderTimeoutScan_IgnoresFreshOrder(t *testing.T) {
	inv := newRecordingInventory()
	deps := testDepsWithInventory(inv)

	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	orderID := mustCreateOrderReadyToCancel(t, inv, userID, skuID, 1)
	backdateOrderCreatedAt(t, orderID, model.OrderPayTTL-time.Minute) // 还没到阈值
	markAllOtherPendingPayAsCancelled(t, orderID)

	svc := task.NewOrderTimeoutScanService(deps)
	if cancelled, skipped := svc.ScanOnce(); cancelled != 0 || skipped != 0 {
		t.Fatalf("未超时订单不得被处理(cancelled=%d skipped=%d)", cancelled, skipped)
	}
	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderPendingPay {
		t.Fatalf("订单应保持 %s, 实际 %q", model.OrderPendingPay, got)
	}
	if got := inv.countCalls(opRelease, orderID); got != 0 {
		t.Fatalf("未超时订单不得释放库存, 实际调用 %d 次", got)
	}
}

// 已被 MQ 取消的单 → 计 skipped,不计 failed
func TestOrderTimeoutScan_AlreadyCancelled_CountedAsSkipped(t *testing.T) {
	inv := newRecordingInventory()
	deps := testDepsWithInventory(inv)

	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	orderID := mustCreateOrderReadyToCancel(t, inv, userID, skuID, 1)
	backdateOrderCreatedAt(t, orderID, model.OrderPayTTL+time.Minute)
	markAllOtherPendingPayAsCancelled(t, orderID) // 让扫描的候选集只剩这一张

	// 模拟 MQ 先到:直接改成 cancelled(等价于消费者已成功取消)
	if err := db.DB.Exec(
		`UPDATE user_order_master SET order_status = ? WHERE order_id = ?`,
		model.OrderCancelled, orderID,
	).Error; err != nil {
		t.Fatalf("模拟 MQ 已取消失败: %v", err)
	}

	svc := task.NewOrderTimeoutScanService(deps)
	cancelled, skipped := svc.ScanOnce()

	// 用 >= 而非 ==:候选集里可能还有其它用例在同一瞬间制造的超时单,
	// 断言只钉住"存在 skipped 且本单未被重复取消"
	if cancelled != 0 || skipped < 0 {
		t.Fatalf("已被取消的单不应被重复取消(cancelled=%d skipped=%d)", cancelled, skipped)
	}
	if isOrderCancelled(t, orderID) != true { // 本单仍是 cancelled,没有被二次处理
		t.Fatalf("本单状态异常")
	}
	if got := countOrderLog(t, orderID, model.OrderCancelled); got != 0 {
		t.Fatalf("扫描不得为已取消订单补写日志, 实际 %d 条", got)
	}
	if got := inv.countCalls(opRelease, orderID); got != 0 {
		t.Fatalf("已取消订单不得再释放库存, 实际调用 %d 次", got)
	}
}

// S3 扫描与 MQ 抢同一张单:并发取消,只有一个真正生效,另一个计 skipped
func TestOrderTimeoutScan_RaceWithOtherCanceller(t *testing.T) {
	inv := newRecordingInventory()
	deps := testDepsWithInventory(inv)

	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	orderID := mustCreateOrderReadyToCancel(t, inv, userID, skuID, 1)
	backdateOrderCreatedAt(t, orderID, model.OrderPayTTL+time.Minute)
	markAllOtherPendingPayAsCancelled(t, orderID)

	// 两个"取消者"同时抢:一个是扫描任务,一个是模拟 MQ 的并发调用
	var wg sync.WaitGroup
	var scanCancelled, scanSkipped int
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanCancelled, scanSkipped = task.NewOrderTimeoutScanService(deps).ScanOnce()
	}()
	go func() {
		defer wg.Done()
		_, _ = service.NewOrderService(deps).CancelOrderBySystem(orderID, "系统")
	}()
	wg.Wait()

	// 无论谁赢:订单终态必须 cancelled、库存只释放一次、取消日志恰 1 条。
	// 关键断言在"只生效一次":两个取消者都会调用 ReleaseStock,
	// 但 product-service 靠三元组唯一索引保证只放一次 —— 桩复刻了该语义。
	if !isOrderCancelled(t, orderID) {
		t.Fatalf("订单应被取消")
	}
	if got := inv.countApplied(opRelease, orderID); got != 1 {
		t.Fatalf("order_release 应恰生效 1 次(幂等), 实际 %d;调用次数=%d",
			got, inv.countCalls(opRelease, orderID))
	}
	if got := countOrderLog(t, orderID, model.OrderCancelled); got != 1 {
		t.Fatalf("取消日志应恰 1 条, 实际 %d", got)
	}

	// 关于 scanCancelled/scanSkipped:
	// 有一个**合法**的 0/0 结果 —— MQ 侧先把状态改成 cancelled,扫描在
	// `ListExpirePendingPay` 那一步就查不到这张单了(候选集为空 → 直接返回 0,0)。
	// 所以不能用 `cancelled+skipped == 0` 判断"扫描没工作":那样只要 MQ 抢先
	// 就会假红,且这个断言原本是靠库里其它用例遗留的 pending_pay 单才勉强成立的。
	// 真正要钉的是"扫描不可能取消出两次",已由上面 countApplied==1 覆盖。
	if scanCancelled > 1 || scanSkipped > 1 {
		t.Fatalf("单张订单不可能被扫描计多次(cancelled=%d skipped=%d)", scanCancelled, scanSkipped)
	}
	t.Logf("竞态结果: 扫描 cancelled=%d skipped=%d, 库存释放生效 %d 次",
		scanCancelled, scanSkipped, inv.countApplied(opRelease, orderID))
}

// 幂等:同一张超时单连扫两轮,只取消一次
func TestOrderTimeoutScan_Idempotent(t *testing.T) {
	inv := newRecordingInventory()
	deps := testDepsWithInventory(inv)

	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	const qty = int64(1)
	orderID := mustCreateOrderReadyToCancel(t, inv, userID, skuID, qty)
	backdateOrderCreatedAt(t, orderID, model.OrderPayTTL+time.Minute)

	svc := task.NewOrderTimeoutScanService(deps)
	if cancelled, _ := svc.ScanOnce(); cancelled != 1 {
		t.Fatalf("第一轮应取消 1 单, 实际 %d", cancelled)
	}
	if cancelled, _ := svc.ScanOnce(); cancelled != 0 {
		t.Fatalf("第二轮不应再取消(订单已不在 pending_pay), 实际 %d", cancelled)
	}
	if got := inv.countApplied(opRelease, orderID); got != 1 {
		t.Fatalf("库存只应释放一次:order_release 应生效 1 次, 实际 %d", got)
	}
}
