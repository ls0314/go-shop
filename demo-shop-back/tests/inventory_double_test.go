package tests

import (
	"fmt"
	"sync"

	"demo-shop-back/src/service"
)

// ============================================================
// 库存 RPC 测试桩(阶段 C2)
//
// 库存四操作与手动调整已迁 product-service,状态存在 product_db,
// 不再能从测试库直接读 sys_product_sku 断言。测试里真正要钉住的也不是
// "product-service 内部算对了没"(那是它自己的并发用例的职责),而是
// **订单域有没有按 (sku, qty, orderId) 正确地调用库存接口、且不重复生效**:
// 取消单要释放、重复取消只释放一次、并发取消只有一个真正生效。
//
// 因此桩记录调用并模拟 product-service 的幂等:同一 (op, orderId, skuId)
// 只生效一次 —— 与真实实现依赖的 sys_product_stock_log 三元组唯一索引同一语义。
// ============================================================

// stockOp 一次库存操作记录
type stockOp struct {
	Op      string // lock / release / deduct / refund
	SkuId   int64
	Qty     int64
	OrderId int64
}

// 操作类型常量(与库表 change_type 同名,便于对照)
const (
	opLock    = "order_lock"
	opRelease = "order_release"
	opDeduct  = "pay_deduct"
	opRefund  = "refund_release"
)

// recordingInventory 实现 service.InventoryStockRPC
type recordingInventory struct {
	mu      sync.Mutex
	ops     []stockOp
	applied map[string]bool // 幂等账本:(op|orderId|skuId)
}

func newRecordingInventory() *recordingInventory {
	return &recordingInventory{applied: map[string]bool{}}
}

var _ service.InventoryStockRPC = (*recordingInventory)(nil)

// record 记录调用并按三元组去重,返回本次是否"真正生效"。
// 与 product-service 的语义一致:重复调用返回成功(幂等命中),但不重复扣/放。
func (r *recordingInventory) record(op string, skuId, qty, orderId int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s|%d|%d", op, orderId, skuId)
	if r.applied[key] {
		// 幂等命中:仍记一笔调用,便于断言"确实被重复调用过"
		r.ops = append(r.ops, stockOp{Op: op + "#idempotent", SkuId: skuId, Qty: qty, OrderId: orderId})
		return nil
	}
	r.applied[key] = true
	r.ops = append(r.ops, stockOp{Op: op, SkuId: skuId, Qty: qty, OrderId: orderId})
	return nil
}

func (r *recordingInventory) LockStock(skuId, qty, orderId int64) error {
	return r.record(opLock, skuId, qty, orderId)
}

func (r *recordingInventory) DeductStock(skuId, qty, orderId int64) error {
	return r.record(opDeduct, skuId, qty, orderId)
}

func (r *recordingInventory) ReleaseStock(skuId, qty, orderId int64) error {
	return r.record(opRelease, skuId, qty, orderId)
}

func (r *recordingInventory) RefundStock(skuId, qty, orderId int64) error {
	return r.record(opRefund, skuId, qty, orderId)
}

// countApplied 统计某订单上"真正生效"的某类操作次数。
// 断言"库存只释放一次"用这个 —— 幂等命中会留下调用记录,但不应重复生效。
func (r *recordingInventory) countApplied(op string, orderId int64) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := 0
	for _, o := range r.ops {
		if o.Op == op && o.OrderId == orderId {
			n++
		}
	}
	return n
}

// countCalls 统计某订单上某类操作的调用次数(含幂等命中)
func (r *recordingInventory) countCalls(op string, orderId int64) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := 0
	for _, o := range r.ops {
		if (o.Op == op || o.Op == op+"#idempotent") && o.OrderId == orderId {
			n++
		}
	}
	return n
}

// findApplied 取某订单上第一次生效的某类操作(断言 sku 与数量)
func (r *recordingInventory) findApplied(op string, orderId int64) (stockOp, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, o := range r.ops {
		if o.Op == op && o.OrderId == orderId {
			return o, true
		}
	}
	return stockOp{}, false
}

// testDepsWithInventory 把桩注入依赖,供订单/支付/扫描任务侧使用。
//
// 桩必须经 deps 传进去:被测代码内部会用 deps 再构造子服务
// (OrderService 内嵌 CartItemService、扫描任务自己 NewOrderService),
// 只塞给某一个结构体会让另一条路径仍拿到 nil。
func testDepsWithInventory(inv *recordingInventory) service.ServiceDeps {
	deps := testDeps()
	deps.InventoryRPC = inv
	return deps
}
