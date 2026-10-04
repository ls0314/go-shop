package productclient

import (
	"context"
	"errors"
	"time"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/service"
)

// ErrUnavailable 客户端未建连(etcd 连不上 / 服务未注册)时的统一错误。
var ErrUnavailable = errors.New("product-service 不可用")

// callTimeout 单次商品/库存 RPC 的超时。
const callTimeout = 5 * time.Second

// Client product-service 客户端,一条连接同时实现两个接口:
//
//	service.InventoryRPC    S3 锁库存 / 补偿释放 / 支付后扣减
//	service.ProductReadRPC  SKU 级读取(购物车回填、下单取价与快照)
type Client struct {
	inventory v1_productv1.InventoryServiceClient
	product   v1_productv1.ProductServiceClient
}

// NewClient 用同一条 gRPC 连接构造两个域的客户端
func NewClient(inventory v1_productv1.InventoryServiceClient, product v1_productv1.ProductServiceClient) *Client {
	return &Client{inventory: inventory, product: product}
}

func (c *Client) ok() bool {
	return c != nil && c.inventory != nil && c.product != nil
}

// ============================================================
// 库存域(service.InventoryRPC)
// ============================================================

// LockStock 锁定库存。orderId 同时是幂等键(同一订单同一 SKU 只锁一次)。
func (c *Client) LockStock(skuId, quantity int64, idempotencyKey string) error {
	if !c.ok() {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.inventory.LockStock(ctx, &v1_productv1.LockStockReq{
		SkuId:          skuId,
		Qty:            quantity,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// ReleaseStock 释放已锁库存(补偿)。幂等:按 (orderId, skuId) 只释放一次。
func (c *Client) ReleaseStock(skuId, quantity int64, idempotencyKey string) error {
	if !c.ok() {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.inventory.ReleaseStock(ctx, &v1_productv1.ReleaseStockReq{
		SkuId:          skuId,
		Qty:            quantity,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// DeductStock 支付成功后把锁定库存转为实际扣减
func (c *Client) DeductStock(skuId, quantity int64, idempotencyKey string) error {
	if !c.ok() {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.inventory.DeductStock(ctx, &v1_productv1.DeductStockReq{
		SkuId:          skuId,
		Qty:            quantity,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// ============================================================
// 商品读域(service.ProductReadRPC)
// ============================================================

// SkuSnapshot 别名到 service 包的类型。
type SkuSnapshot = service.SkuSnapshot

// BatchGetSkus 批量取 SKU 快照。
func (c *Client) BatchGetSkus(skuIds []int64) (map[int64]*service.SkuSnapshot, error) {
	if !c.ok() {
		return nil, ErrUnavailable
	}
	if len(skuIds) == 0 {
		return map[int64]*SkuSnapshot{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.product.BatchGetSkus(ctx, &v1_productv1.BatchGetSkusReq{SkuIds: skuIds})
	if err != nil {
		return nil, err
	}
	if err := RestoreError(resp.ErrorMsg); err != nil {
		return nil, err
	}

	out := make(map[int64]*SkuSnapshot, len(resp.Items))
	for _, it := range resp.Items {
		if it == nil || it.Sku == nil {
			continue
		}
		sku := it.Sku
		out[sku.SkuId] = &SkuSnapshot{
			SkuId:      sku.SkuId,
			SpuId:      sku.SpuId,
			SkuName:    sku.SkuName,
			SpuName:    it.SpuName,
			SpecValues: sku.SpecValues,
			MainImage:  it.SpuMainImage,
			SkuImage:   sku.SkuImage,
			Price:      sku.Price,
			Stock:      sku.Stock,
			Available:  it.SkuActive && !it.SkuDeleted && it.SpuPublished && !it.SpuDeleted,
		}
	}
	return out, nil
}

// RestoreError 把服务端 error_msg 还原成本地哨兵错误(能还原时)或普通错误。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if err, ok := serverErrMap[errorMsg]; ok {
		return err
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// serverErrMap 服务端文案 → 本地哨兵错误。
var serverErrMap = map[string]error{
	"SKU不存在":     model.ErrSkuNotExist,
	"库存不足":       model.ErrStockNotEnough,
	"锁定库存不足":     model.ErrLockStockNotEnough,
	"SKU已禁用或已删除": model.ErrSkuDisabled,
}
