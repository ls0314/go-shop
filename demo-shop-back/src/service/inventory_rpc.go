package service

// InventoryStockRPC 库存四操作的抽象。
//
// 为什么要抽出接口:库存四操作已迁 product-service,`*inventoryclient.InventoryClient`
// 是**具体类型且只能经 etcd 建连**。测试里要验证"取消订单是否按 (sku, qty, orderId)
// 释放了库存",就必须能替换这一层 —— 否则每个用例都要 etcd + 一个跑起来的
// product-service + 指向同一个测试库的配置,代价远超收益(见待办-C2 §8.10/§8.13)。
//
// 生产实现是 `*inventoryclient.InventoryClient`,它天然满足本接口,
// 因此 NewOrderService / NewPaymentService 的签名不必改。
type InventoryStockRPC interface {
	LockStock(skuId, qty, orderId int64) error
	DeductStock(skuId, qty, orderId int64) error
	ReleaseStock(skuId, qty, orderId int64) error
	RefundStock(skuId, qty, orderId int64) error
}
