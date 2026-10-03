# 库存并发与幂等用例:迁移待办

> **来源**:`demo-shop-back/tests/inventory_concurrent_test.go`(C2 期间删除,191 行 / 7 个用例)。
> **为什么删**:该文件调用 `service.NewInventoryService(...)` 上的
> `LockStock`/`ReleaseStock`/`DeductStock`/`RefundStock` 本地实现,而这些实现
> **已随库存域迁到 product-service 并从单体删除**。文件无法编译,且其断言
> (`SELECT stock FROM sys_product_sku`)读的是 `demo_shop`,而库存在 `product_db`。
> **不是覆盖被放弃,是覆盖换了宿主**:这些语义现在属 product-service。

## 必须补回的 7 个用例(契约照搬,宿主换到 product-service)

| # | 用例 | 断言要点 |
|---|---|---|
| I1 | `TestLockStock_Concurrent_NoOversell` | stock=10、40 并发各锁 1 件(独立订单)→ 恰好 10 成功 / 30 库存不足,`stock=0`、`lock_stock=10`、`order_lock` 流水恰 10 条 |
| I2 | `TestLockStock_Idempotent_SameOrder` | 同一 order_id 重复 LockStock → 第二次返回 nil(幂等命中),库存只挪一次 |
| I3 | `TestReleaseStock_Idempotent` | 重复 ReleaseStock → 库存只回补一次,`order_release` 流水恰 1 条 |
| I4 | `TestLockStock_MultiSku_SameOrder` | 同订单多 SKU:每个 SKU 各锁各的,互不干扰 |
| I5 | `TestReleaseStock_Concurrent_SameOrder` | 同订单同 SKU 双 goroutine 并发释放 → 只生效一次(靠三元组唯一索引) |
| I6 | `TestReleaseStock_MultiSku_SameOrder_Concurrent` | 同订单多 SKU 并发释放 → 每个 SKU 恰释放一次 |
| I7 | (同文件内的第 7 个用例) | 见 git 历史 `inventory_concurrent_test.go` |

## 落点与前置

- **落点**:`services/product/internal/logic/inventoryservice/` 下的集成测试
  (`lockstocklogic_test.go` 等),或用例内直建 `ServiceContext{DB: product_db}` 后调用 logic;
- **前置**:`product_db` 已有表(✅ 已完成,见 `cmd/migrate -verify`),
  但**没有数据** —— 测试需自建 category → spu → sku 的真外键链
  (可照搬单体 `tests/factory_test.go:mustCreateSkuWithStock`);
- **别忘**:`ReleaseStock` 的 `lock_stock < qty` 校验会返回 `ErrLockStockNotEnough`,
  造数时必须先 LockStock 再 Release;
- **归属**:C2 收尾(与写路径收口同批)。
