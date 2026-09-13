package service

import (
	"errors"

	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// LockStockWithTx 锁库存入口(带业务指标埋点,DS-A-22)。
// 超卖拒绝率(stock_lock_total{result=not_enough})是库存健康度的直接观测点
func (is *InventoryService) LockStockWithTx(tx *gorm.DB, skuId, qty, orderId int64) error {
	err := is.lockStockWithTx(tx, skuId, qty, orderId)

	result := "error"
	switch {
	case err == nil:
		result = "success" // 含幂等跳过(重复消息)——语义上仍是「本次锁定诉求已满足」
	case errors.Is(err, model.ErrStockNotEnough):
		result = "not_enough"
	}
	metrics.StockLockTotal.WithLabelValues(result).Inc()
	return err
}
