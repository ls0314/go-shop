package inventoryclient

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop/api/gen/product/v1"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// stockCallTimeout 单次库存 RPC 的超时。与 userclient 保持同一取舍:
// 用 context.Background() 而非透传请求 ctx,客户端断开后 RPC 仍会跑完
// (最多这个时长),避免库存操作被中途取消留下状态不明。
const stockCallTimeout = 2 * time.Second

// InventoryClient 库存四操作的 RPC 客户端。
//
// 这四个操作在 product-service 侧自带事务与幂等(流水表
// (sku_id, order_id, change_type) 唯一索引),调用方可安全重试。
type InventoryClient struct {
	inventory v1_productv1.InventoryServiceClient
	conn      *grpc.ClientConn
}

// NewInventoryClient 建连 product-service(etcd 服务发现)。
func NewInventoryClient(etcdHosts []string, etcdKey string) (*InventoryClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	return &InventoryClient{
		inventory: v1_productv1.NewInventoryServiceClient(client.Conn()),
		conn:      client.Conn(),
	}, nil
}

func (c *InventoryClient) Close() error { return c.conn.Close() }

// LockStock 下单锁定库存。失败时返回的 error 已还原为单体 model 包的
// 错误变量,调用方可用 errors.Is 判定。
func (c *InventoryClient) LockStock(skuId, qty, orderId int64) error {
	if c == nil {
		return errors.New("product-service 不可用")
	}
	ctx, cancel := context.WithTimeout(context.Background(), stockCallTimeout)
	defer cancel()

	resp, err := c.inventory.LockStock(ctx, &v1_productv1.LockStockReq{
		SkuId:   skuId,
		Qty:     qty,
		OrderId: orderId,
	})
	if err != nil {
		return err
	}
	return toDomainError(resp.ErrorMsg)
}

// DeductStock 支付成功扣减库存。调用方为支付链路,
// 失败不做反向补偿(支付不可逆),由调用方重试 + 对账兜底。
func (c *InventoryClient) DeductStock(skuId, qty, orderId int64) error {
	if c == nil {
		return errors.New("product-service 不可用")
	}
	ctx, cancel := context.WithTimeout(context.Background(), stockCallTimeout)
	defer cancel()

	resp, err := c.inventory.DeductStock(ctx, &v1_productv1.DeductStockReq{
		SkuId:   skuId,
		Qty:     qty,
		OrderId: orderId,
	})
	if err != nil {
		return err
	}
	return toDomainError(resp.ErrorMsg)
}

// ReleaseStock 取消订单释放库存。幂等:重复调用只释放一次。
func (c *InventoryClient) ReleaseStock(skuId, qty, orderId int64) error {
	if c == nil {
		return errors.New("product-service 不可用")
	}
	ctx, cancel := context.WithTimeout(context.Background(), stockCallTimeout)
	defer cancel()

	resp, err := c.inventory.ReleaseStock(ctx, &v1_productv1.ReleaseStockReq{
		SkuId:   skuId,
		Qty:     qty,
		OrderId: orderId,
	})
	if err != nil {
		return err
	}
	return toDomainError(resp.ErrorMsg)
}

// RefundStock 退款回补库存。
func (c *InventoryClient) RefundStock(skuId, qty, orderId int64) error {
	if c == nil {
		return errors.New("product-service 不可用")
	}
	ctx, cancel := context.WithTimeout(context.Background(), stockCallTimeout)
	defer cancel()

	resp, err := c.inventory.RefundStock(ctx, &v1_productv1.RefundStockReq{
		SkuId:   skuId,
		Qty:     qty,
		OrderId: orderId,
	})
	if err != nil {
		return err
	}
	return toDomainError(resp.ErrorMsg)
}

// toDomainError 把 RPC 的 error_msg 还原为单体 model 包的错误变量。
//
// 为什么必须还原而不能直接返回字符串:调用方与埋点用
// errors.Is(err, model.ErrXxx) 判定错误类型(inventory_metrics.go:21),
// 返回新构造的 error 会让判定恒为 false,指标与错误码静默失效。
//
// 文案是跨服务的契约,product-service 侧修改文案必须同步这里。
func toDomainError(errorMsg string) error {
	switch errorMsg {
	case "":
		return nil
	case model.ErrSkuNotExist.Error():
		return model.ErrSkuNotExist
	case model.ErrSkuDisabled.Error():
		return model.ErrSkuDisabled
	case model.ErrStockNotEnough.Error():
		return model.ErrStockNotEnough
	case model.ErrLockStockNotEnough.Error():
		return model.ErrLockStockNotEnough
	case model.ErrStockNegative.Error():
		return model.ErrStockNegative
	default:
		return errors.New(errorMsg)
	}
}
