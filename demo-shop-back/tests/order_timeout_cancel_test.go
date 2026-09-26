package tests

import (
	"errors"
	"testing"
	"time"

	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"gorm.io/datatypes"
)

// ============================================================
// 订单超时取消 / 系统取消 测试(DS-A-24 A2)
//
// 背景:改前 MQ 消费者调用 `canceller.CancelOrder(orderId, 4, "系统")`——userId 被写死成
// 常量 4,而 CancelOrder 内部有归属校验(userId != order.UserId → ErrOrderNoPermission),
// 于是**除 user_id=4 之外的所有订单都无法被自动取消**:超时消息落到 failed 分支后
// 被 msg.Ack(false) 丢弃,订单永久停在 pending_pay、锁定库存永久不回补。
//
// 本文件是该缺陷的回归测试:系统取消不携带 userId,归属校验只保留在用户主动取消路径上。
// ============================================================

// mustCreatePendingOrder 直插一张 pending_pay 订单(不经过 CreateOrder 的完整链路:
// 本用例要测的是取消路径,下单链路的正确性由库存/券的并发用例各自覆盖)。
//
// order_id 显式生成(复用 nextOrderID),使它与本轮其它用例的订单互不干扰;
// 注意 user_order_master.order_id 是 BIGSERIAL,显式赋值不会推进序列,
// 与 inventory_concurrent_test.go 里"order_id 无外键,可编"的既有做法一致。
func mustCreatePendingOrder(t *testing.T, userId int64, skuId, qty int64) int64 {
	t.Helper()

	orderId := nextOrderID()
	unitPrice := 10.0
	total := unitPrice * float64(qty)
	now := time.Now()

	o := model.UserOrder{
		OrderId:         orderId,
		OrderNo:         "TESTNO-" + uniqueTag(),
		UserId:          userId,
		OrderStatus:     model.OrderPendingPay,
		TotalAmount:     total,
		PayAmount:       total, // 等于 TotalAmount ⇒ 取消时不走退券分支,聚焦库存路径
		AddressSnapshot: []byte(`{}`),
		BuyerRemark:     "timeout-cancel-test",
		IdempotentKey:   "TESTIDEM-" + uniqueTag(), // NOT NULL + 唯一索引,必须给唯一值
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
		DetailCount:     1,
		FirstImage:      "http://example.invalid/test.png", // NOT NULL
	}

	if err := db.DB.Select(
		"order_id", "order_no", "user_id", "order_status", "total_amount", "pay_amount",
		"address_snapshot", "buyer_remark", "idempotent_key", "is_deleted",
		"created_at", "updated_at", "detail_count", "first_image",
	).Create(&o).Error; err != nil {
		t.Fatalf("造订单失败: %v", err)
	}
	return orderId
}

func mustCreateOrderDetail(t *testing.T, orderId, skuId, qty int64) {
	t.Helper()

	d := model.UserOrderDetail{
		OrderId:    orderId,
		SkuId:      skuId,
		SpuName:    "TESTSPU",
		SkuName:    "TESTSKU",
		MainImage:  "http://example.invalid/test.png",
		Quantity:   qty,
		UnitPrice:  10,
		TotalPrice: 10 * float64(qty),
		CreatedAt:  time.Now(),
		SpecValues: datatypes.JSONMap{},
	}
	if err := db.DB.Select(
		"order_id", "sku_id", "spu_name", "sku_name", "spec_values", "main_image",
		"quantity", "unit_price", "total_price", "created_at",
	).Create(&d).Error; err != nil {
		t.Fatalf("造订单明细失败: %v", err)
	}
}

// mustCreateOrderReadyToCancel 造一张「真实下单后」的订单:主表 + 明细 + 库存已锁定。
// 三个动作必须一起做——只建单不锁库存会得到 "lock_stock=0 但明细要释放 1 件" 的畸形状态,
// 取消时 ReleaseStockWithTx 的 `lock_stock < qty` 校验会返回 ErrStockNegative
// (inventory_service.go:527),与真实下单流程不符。
func mustCreateOrderReadyToCancel(t *testing.T, userId, skuId, qty int64) int64 {
	t.Helper()

	orderID := mustCreatePendingOrder(t, userId, skuId, qty)
	mustCreateOrderDetail(t, orderID, skuId, qty)
	if err := service.NewInventoryService().LockStock(skuId, qty, orderID); err != nil {
		t.Fatalf("前置锁库存失败: %v", err)
	}
	return orderID
}

// O1 超时系统取消(核心回归):订单属于非 4 号用户时,系统取消必须成功。
//
// 这条断言在改前必然失败——旧接口没有任何办法在不伪造 userId 的情况下取消别人的订单,
// 而消费者伪造的那个值(4)只对 user_id=4 的订单有效。因此本用例钉住的正是"定值 userId"
// 这个缺陷本身,而不是它的某一种表现。
func TestCancelOrderBySystem_AnyUser_Succeeds(t *testing.T) {
	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	const qty = int64(2)
	orderID := mustCreateOrderReadyToCancel(t, userID, skuID, qty)

	svc := service.NewOrderService()
	resp, err := svc.CancelOrderBySystem(orderID, "系统")
	if err != nil {
		t.Fatalf("系统取消应当成功(订单属于 user_id=%d,与旧代码写死的 4 无关), 实际报错: %v",
			userID, err)
	}
	if resp == nil || resp.OrderStatus != model.OrderCancelled {
		t.Fatalf("取消后状态应为 %s, 实际 %+v", model.OrderCancelled, resp)
	}

	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderCancelled {
		t.Fatalf("订单状态应为 %s, 实际 %q", model.OrderCancelled, got)
	}
	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 10 {
		t.Fatalf("库存应回补到 10, 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 0 {
		t.Fatalf("锁定库存应归零, 实际 %d", got)
	}
	if got := queryInt64(t,
		`SELECT COUNT(*) FROM sys_product_stock_log WHERE order_id = ? AND sku_id = ? AND change_type = ?`,
		orderID, skuID, model.StockOrderRelease); got != 1 {
		t.Fatalf("order_release 流水应为 1 条, 实际 %d", got)
	}
	if got := queryInt64(t,
		`SELECT COUNT(*) FROM user_order_log WHERE order_id = ? AND order_status = ?`,
		orderID, model.OrderCancelled); got != 1 {
		t.Fatalf("订单日志应记录 1 条取消操作, 实际 %d", got)
	}
	if got := queryString(t, `SELECT operator FROM user_order_log WHERE order_id = ?`, orderID); got != "系统" {
		t.Fatalf("订单日志操作人应为「系统」, 实际 %q", got)
	}
}

// O2 系统取消幂等:重复取消不得重复释放库存。
// MQ 消息重复投递 + A2b 扫描任务并发命中同一张单是**必然会发生**的场景
// (outbox 是 at-least-once,见 DS-A-23 5.1),所以取消路径必须可重复调用。
func TestCancelOrderBySystem_Idempotent(t *testing.T) {
	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	const qty = int64(3)
	orderID := mustCreateOrderReadyToCancel(t, userID, skuID, qty)

	svc := service.NewOrderService()
	if _, err := svc.CancelOrderBySystem(orderID, "系统"); err != nil {
		t.Fatalf("首次系统取消失败: %v", err)
	}
	// 第二次:订单已 cancelled → 状态机返回 ErrOrderCannotCancel。
	// 这是**正常跳过**,不是故障:调用方(消费者/扫描任务)应把它当幂等命中处理。
	if _, err := svc.CancelOrderBySystem(orderID, "系统"); !errors.Is(err, model.ErrOrderCannotCancel) {
		t.Fatalf("重复取消应返回 ErrOrderCannotCancel(幂等跳过), 实际: %v", err)
	}

	if got := queryInt64(t, `SELECT stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 10 {
		t.Fatalf("库存只应回补一次(10), 实际 %d", got)
	}
	if got := queryInt64(t, `SELECT lock_stock FROM sys_product_sku WHERE sku_id = ?`, skuID); got != 0 {
		t.Fatalf("锁定库存应归零, 实际 %d", got)
	}
	if got := queryInt64(t,
		`SELECT COUNT(*) FROM sys_product_stock_log WHERE order_id = ? AND change_type = ?`,
		orderID, model.StockOrderRelease); got != 1 {
		t.Fatalf("重复取消不得重复释放:order_release 应为 1 条, 实际 %d", got)
	}
	if got := queryInt64(t,
		`SELECT COUNT(*) FROM user_order_log WHERE order_id = ? AND order_status = ?`,
		orderID, model.OrderCancelled); got != 1 {
		t.Fatalf("订单日志应只有 1 条取消记录, 实际 %d", got)
	}
}

// O3 不取消已支付订单(保护):支付回调与超时取消存在天然竞态——用户在第 15 分钟付款时,
// 延迟消息可能已经到期。系统取消必须拒绝已支付订单,否则会"取消掉付过钱的单"。
func TestCancelOrderBySystem_PaidOrder_Rejected(t *testing.T) {
	userID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	orderID := mustCreateOrderReadyToCancel(t, userID, skuID, 1)
	// 再 UPDATE 成 paid
	if err := db.DB.Exec(
		`UPDATE user_order_master SET order_status = ? WHERE order_id = ?`,
		model.OrderPaid, orderID,
	).Error; err != nil {
		t.Fatalf("模拟支付失败: %v", err)
	}

	svc := service.NewOrderService()
	if _, err := svc.CancelOrderBySystem(orderID, "系统"); !errors.Is(err, model.ErrOrderCannotCancel) {
		t.Fatalf("已支付订单不得被系统取消,期望 ErrOrderCannotCancel, 实际: %v", err)
	}
	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderPaid {
		t.Fatalf("订单状态必须保持 %s, 实际 %q", model.OrderPaid, got)
	}
}

// O4 用户主动取消仍保留归属校验:系统路径放开归属校验,但**不能顺带削弱用户路径**——
// 否则任何登录用户都能取消别人的订单。
func TestCancelOrder_UserPath_OwnershipStillEnforced(t *testing.T) {
	ownerID := mustCreateUser(t)
	otherID := mustCreateUser(t)
	skuID, _ := mustCreateSkuWithStock(t, 10)

	orderID := mustCreateOrderReadyToCancel(t, ownerID, skuID, 1)

	svc := service.NewOrderService()
	// 别人来取消 → 必须被拒(旧代码里消费者伪造 userId=4 绕过校验的做法,在用户路径上依然无效)
	if _, err := svc.CancelOrder(orderID, otherID, "冒名者"); !errors.Is(err, model.ErrOrderNoPermission) {
		t.Fatalf("非订单归属人取消应返回 ErrOrderNoPermission, 实际: %v", err)
	}
	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderPendingPay {
		t.Fatalf("越权取消不得改动订单状态, 实际 %q", got)
	}
	// 本人取消 → 放行
	if _, err := svc.CancelOrder(orderID, ownerID, "本人"); err != nil {
		t.Fatalf("订单归属人取消应成功, 实际: %v", err)
	}
	if got := queryString(t, `SELECT order_status FROM user_order_master WHERE order_id = ?`, orderID); got != model.OrderCancelled {
		t.Fatalf("本人取消后状态应为 %s, 实际 %q", model.OrderCancelled, got)
	}
}
