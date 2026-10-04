package service

// ============================================================
// 跨服务调用的客户端接口
// ============================================================

// InventoryRPC 库存域(锁库存 / 补偿释放 / 支付后扣减)。
type InventoryRPC interface {
	// LockStock 锁定库存
	LockStock(skuId, quantity int64, idempotencyKey string) error
	// ReleaseStock 释放已锁库存(补偿)。按 (幂等键, skuId, change_type) 幂等,可重放
	ReleaseStock(skuId, quantity int64, idempotencyKey string) error
	// DeductStock 支付成功后把锁定库存转为实际扣减
	DeductStock(skuId, quantity int64, idempotencyKey string) error
}

// CouponRPC 券域(核销 / 补偿退还)。
type CouponRPC interface {
	// UseCoupon 核销券,返回用券后的实付金额。
	// 幂等键是 idempotencyKey;orderNo 只是追溯字段(这张券用在哪张单上)
	UseCoupon(userCouponId int64, idempotencyKey, orderNo string, userId int64, orderAmount float64) (float64, error)
	// ReturnCoupon 按幂等键归还券(补偿)。
	// 幂等:该键没核销过、或券已归还,都算成功
	ReturnCoupon(idempotencyKey string, userId int64) error
}

// ProductReadRPC 商品读域(SKU 级读取)。
//
// 购物车回填展示字段、下单固化商品快照与算价都要用它。
// 与 InventoryRPC 是同一个服务的不同 gRPC service,生产上由
// infra/productclient 的同一个对象实现(一条连接两个域)。
type ProductReadRPC interface {
	// BatchGetSkus 批量取 SKU 快照。返回 map 而非切片:调用方都是
	// "按 skuId 找"的用法,切片会逼它自己建一遍索引。
	BatchGetSkus(skuIds []int64) (map[int64]*SkuSnapshot, error)
}

// SkuSnapshot SKU 快照,订单域真正需要的字段子集。
type SkuSnapshot struct {
	SkuId   int64
	SpuId   int64
	SkuName string
	SpuName string
	// SpecValues 规格值的 JSON 文本,与 sys_product_sku.spec_values 一致
	SpecValues string
	// MainImage SPU 主图(列表页优先用它)
	MainImage string
	// SkuImage SKU 自己的图,可能为空
	SkuImage string
	Price    float64
	Stock    int64
	// Available SKU/SPU 是否可购买(active + published + 未删除)
	Available bool
}

// ============================================================
// 测试用的最小桩
// ============================================================

// InventoryFunc 用函数字段实现 InventoryRPC。未设的字段是 no-op。
type InventoryFunc struct {
	LockFn    func(skuId, quantity int64, idempotencyKey string) error
	ReleaseFn func(skuId, quantity int64, idempotencyKey string) error
	DeductFn  func(skuId, quantity int64, idempotencyKey string) error
}

func (f InventoryFunc) LockStock(skuId, quantity int64, idempotencyKey string) error {
	if f.LockFn == nil {
		return nil
	}
	return f.LockFn(skuId, quantity, idempotencyKey)
}

func (f InventoryFunc) ReleaseStock(skuId, quantity int64, idempotencyKey string) error {
	if f.ReleaseFn == nil {
		return nil
	}
	return f.ReleaseFn(skuId, quantity, idempotencyKey)
}

func (f InventoryFunc) DeductStock(skuId, quantity int64, idempotencyKey string) error {
	if f.DeductFn == nil {
		return nil
	}
	return f.DeductFn(skuId, quantity, idempotencyKey)
}

// CouponFunc 用函数字段实现 CouponRPC。未设的字段是 no-op
// (UseCoupon 的 no-op 返回原金额,即"不用券")。
type CouponFunc struct {
	UseFn    func(userCouponId int64, orderNo string, userId int64, orderAmount float64) (float64, error)
	ReturnFn func(orderNo string, userId int64) error
}

func (f CouponFunc) UseCoupon(userCouponId int64, orderNo string, userId int64, orderAmount float64) (float64, error) {
	if f.UseFn == nil {
		return orderAmount, nil
	}
	return f.UseFn(userCouponId, orderNo, userId, orderAmount)
}

func (f CouponFunc) ReturnCoupon(orderNo string, userId int64) error {
	if f.ReturnFn == nil {
		return nil
	}
	return f.ReturnFn(orderNo, userId)
}
