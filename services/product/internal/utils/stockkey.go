package utils

import "fmt"

// StockGateKey 库存闸门键。product-service 与单体共用同一 Redis 实例
func StockGateKey(skuId int64) string {
	return fmt.Sprintf("sku:stock:%d", skuId)
}
