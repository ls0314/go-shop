# TESTDATA-订单域用例待迁

单体 `demo-shop-back/tests/` 里与订单域相关的用例,随订单三表所有权迁出
(C4)**必须移到这里**。迁移前它们在单体里已无法运行:订单表已落
`trade_db`,单体测的 `demo_shop` 里那张表停更了。

从单体删除的文件:

| 文件 | 行数 | 为什么必须移走 |
|---|---|---|
| `order_timeout_cancel_test.go` | ~262 | 调 `service.NewOrderService(deps)` 做本地取消,该方法已删 |
| `order_timeout_scan_test.go` | ~210 | 调 `task.NewOrderTimeoutScanService(deps)`,该任务已删 |

同时删除的夹具(在 `order_timeout_cancel_test.go` 内定义,被上面两个文件共用):
`mustCreatePendingOrder`、`mustCreateOrderDetail`、`mustCreateOrderReadyToCancel`、
`backdateOrderCreatedAt`、`markAllOtherPendingPayAsCancelled`、`isOrderCancelled`、
`countOrderLog`。

---

## 一、超时取消(原 `order_timeout_cancel_test.go`)

### T1 `TestCancelOrder_SystemReleasesStock` —— 系统取消要释放库存

- **场景**:造一张「待支付 + 库存已锁」的订单,调系统取消
- **断言**:订单状态 → `cancelled`;`ReleaseStock` 恰好**真正生效 1 次**;
  订单日志恰好 1 条 `cancel`,且 `operator = "系统"`
- **迁移落点**:`internal/service/order_test.go`(用 `InventoryFunc` 桩)
- **关键**:释放用的幂等键必须与下单锁库存时**同一个**(订单上的 `idempotent_key`)。
  这个键对不上时 `countApplied` 会返回 0,用例会假红 —— 单体验收时就踩过这个坑

### T2 `TestCancelOrder_Idempotent` —— 重复取消只释放一次

- **场景**:连续两次系统取消
- **断言**:第二次返回 `ErrOrderCannotCancel`(或 `ErrOrderAlreadyCancelled`);
  `ReleaseStock` 的**生效次数仍为 1**;`LockStock` 的调用次数不变;取消日志仍为 1 条
- **迁移落点**:同上
- **注意**:断言要盯 `countApplied`(真正生效)而不是 `countCalls`(含幂等命中)——
  补偿会被重放,幂等命中是**正常**的,不是错误

### T3 `TestCancelOrder_PaidOrderRejected` —— 已支付不可取消

- **场景**:把订单改成 `paid` 后调取消
- **断言**:返回 `ErrOrderCannotCancel`;状态仍为 `paid`;`ReleaseStock` **零调用**
- **迁移落点**:同上

### T4 `TestCancelOrder_NoPermission` —— 越权取消被拒

- **场景**:用户 A 的订单,用用户 B 的 ID 取消
- **断言**:返回 `ErrOrderNoPermission`;状态不变;`ReleaseStock` 零调用
- **迁移落点**:同上
- **注意**:系统取消(`user_id = 0`)**不走这条校验** —— 要另有一条用例覆盖
  "user_id 为 0 时不校验归属"

## 二、超时扫描(原 `order_timeout_scan_test.go`)

### T5 `TestOrderTimeoutScan_CancelsExpired` —— 扫描收敛过期单

- **场景**:一张单的 `expire_at` 已过(把它改到阈值之后),跑一轮扫描
- **断言**:该单被取消;`ReleaseStock` 生效 1 次
- **迁移落点**:`internal/task/ordertimeout_test.go`
- **关键**:判据是**订单表的 `expire_at` 列**,不是"创建时间 + 阈值"。
  单体时代这个阈值被三处各自解释(两处 15 分钟、注释写 2 分钟),
  迁出后收敛为"下单时算一次写进 expire_at"。用例要按新口径造数据

### T6 `TestOrderTimeoutScan_SkipsNotYetExpired` —— 未到期不取消

- **场景**:`expire_at` 尚未到,跑一轮扫描
- **断言**:状态仍 `pending_pay`;`ReleaseStock` 零调用
- **迁移落点**:同上

### T7 `TestOrderTimeoutScan_RaceWithOtherCanceller` —— 与其它取消者竞态

- **场景**:先手工把订单改成 `cancelled`(模拟 MQ 抢先),再跑扫描
- **断言**:终态仍是 `cancelled`;`ReleaseStock` **生效恰好 1 次**;
  取消日志恰好 1 条;`scanCancelled <= 1 && scanSkipped <= 1`
- **迁移落点**:同上
- **这条用例原先写错过**:它曾断言"扫描至少命中一次(cancelled + skipped > 0)",
  而 MQ 抢先取消后 `ListExpirePendingPay` 正确地返回空 → 扫描返回 (0,0) 也**是对的**。
  改断言后要钉的是真实不变量(终态 + 补偿只生效一次 + 日志只一条),
  而不是"扫描必须动过手"

## 三、下单 Saga(单体没有的,迁出后**新增**)

单体的下单是单事务,没有补偿路径;改成 Saga 后这几条是新的一等公民:

### S1 补偿逆序且只补已完成的步骤

- **场景**:三步 `券核销 → 锁库存 → 建单`,让**第 2 步失败**
- **断言**:补偿序列恰为 `[退券]`(不含建单 —— 它从没执行过)
- **落点**:`internal/service/saga_test.go`
- **这是整个 Saga 骨架唯一能证明其正确的用例**,且不依赖任何外部服务。
  用函数桩构造步骤即可,不需要 DB / RPC

### S2 建单失败 → 释放库存 + 退券(逆序)

- **场景**:第 3 步(本地事务)失败
- **断言**:先 `ReleaseStock` 后 `ReturnCoupon`;两者都用**同一个幂等键**

### S3 券核销失败 → 什么都不补

- **场景**:第 1 步失败
- **断言**:`ReleaseStock` 与 `ReturnCoupon` 均零调用

### S4 并发重复提交只建一单

- **场景**:同一个 `idempotency_key` 并发两次下单
- **断言**:只落一张单;第二次返回已有订单而不是报错
- **落点**:`internal/service/createorder_test.go`
- **关键**:这靠 `uk_idempotent_key` 唯一索引兜底 + 把 23505 翻译成
  "返回已有订单"(`translateDuplicateKey`)。预检查只是优化,挡不住并发

### S5 支付后扣库存失败 → 不回滚支付(向前补偿)

- **场景**:`DeductStock` 返回错误
- **断言**:支付流水仍 `success`、订单仍 `paid`(钱不能退),
  但返回错误让渠道重试
- **落点**:`internal/service/payment_test.go`
- **这条钉的是 Saga 的**不对称原则**:有不可逆锚点的链路向前补偿

---

## 迁移时需要的夹具

| 单体夹具 | 迁移后在 trade 的形态 |
|---|---|
| `mustCreatePendingOrder` | 直插 `model.UserOrder`(表就在本库),记得填 `expire_at` |
| `mustCreateOrderDetail` | 直插 `model.UserOrderDetail` |
| `mustCreateOrderReadyToCancel` | 直插订单 + 明细,锁库存用 `InventoryFunc` 桩 |
| `backdateOrderCreatedAt` | **改成改 `expire_at`** —— 判据变了,改创建时间不再有效 |
| `markAllOtherPendingPayAsCancelled` | 保留:让扫描候选集只剩目标单,避免用例互相干扰 |
| `newRecordingInventory`(单体 tests/inventory_double_test.go) | trade 侧已有 `service.InventoryFunc`,直接用它 |
| `runConcurrent` / `queryInt64` | 照搬 |
