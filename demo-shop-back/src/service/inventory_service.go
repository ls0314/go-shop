package service

import (
	"demo-shop-back/src/infra/productclient"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
)

// ============================================================
//	定义及实例化
// ============================================================

// InventoryService 库存服务层实例。
//
// 库存域(查询 + 流水 + 四操作 + 手动调整)**全部已迁 product-service**:
//   - 库存看板三个查询 + 流水查询 + 手动调整 → ProductRPC;
//   - 库存四操作 → InventoryRPC(order/payment 服务内直连,不经过本结构)。
//
// 因此本结构不再持有任何 repo:它现在是纯粹的 RPC 门面。
type InventoryService struct {
	ProductRPC   *productclient.ProductClient // 商品/类目/库存 查询 + 手动调整
	InventoryRPC InventoryStockRPC            // 库存四操作(写)
}

// NewInventoryService 创建库存服务层实例
// 接收值：deps - 服务层依赖（由 composition root 注入）
// 返回值：*InventoryService - 库存服务层实例指针
func NewInventoryService(deps ServiceDeps) *InventoryService {
	return &InventoryService{
		ProductRPC:   deps.ProductRPC,
		InventoryRPC: deps.InventoryRPC,
	}
}

// ============================================================
//	库存操作部分(对外接口）
// ============================================================

/* 查询 */

// GetSkuStock 获取单个商品库存信息
// 接收值：skuId - 商品SKU ID
// 返回值：*response.SkuInventoryResp - 商品库存信息详情响应
//
// 读路径已迁 product-service:库存看板要联 sys_product_spu 取 spu_name,
// 而那张表的所有权在 product-service 的库里,本地 JOIN 会跨库。
func (is *InventoryService) GetSkuStock(skuId int64) (*response.SkuInventoryResp, error) {
	stock, errMsg, err := is.ProductRPC.GetSkuStock(skuId)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return stock, nil
}

// GetSkuStockListBySpu 获取spu下商品库存信息列表
// 接收值：spuId - 商品SPU ID
// 返回值：
//
//	*response.SkuInventoryListResp - 商品库存信息列表响应（聚合返回总库存、总销量等信息）
//	error - 错误信息
//
// 聚合值(total_stock/total_lock/total_sold)由服务端累加后返回,本地不再遍历。
func (is *InventoryService) GetSkuStockListBySpu(spuId int64) (*response.SkuInventoryListResp, error) {
	list, errMsg, err := is.ProductRPC.GetSkuStockList(spuId)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return list, nil
}

// GetWarnStockList 获取低于库存阈值的商品库存信息列表
// 接收值：req - 查询参数包含阈值和spu状态信息
// 返回值：
//
//	*response.SkuInventoryListResp - 低于阈值的商品库存信息列表响应（过滤去除内部字段）
//	error - 错误信息
//
// 默认值(threshold 10 / spu_status published)在服务端补,本地不再兜 ——
// 传空值即表示"用服务端默认值",避免两边各写一份默认值后漂移。
func (is *InventoryService) GetWarnStockList(req requset.InventoryWarnReq) (*[]response.InventoryWarnResp, error) {
	list, errMsg, err := is.ProductRPC.GetWarnStockList(req.Threshold, req.SpuStatus)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return &list, nil
}

// GetStockLogList 分页获取库存变更日志列表
// 接收值： req - 查询参数（包含分页、筛选条件等信息）
//
// 返回值：
//
//	*response.InventoryLogResp - 库存变更日志列表
//	error - 错误信息
//
// 已迁 product-service:流水的**写入**早已全在本服务之外的 product-service
// (落在 product_db),此前读 demo_shop 会看到一个停更的副本。
// 分页默认值/封顶(均 50)在服务端补,本地不再兜。
func (is *InventoryService) GetStockLogList(req requset.InventoryLogReq) (*response.InventoryLogResp, error) {
	resp, errMsg, err := is.ProductRPC.ListStockLogs(req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return resp, nil
}

/* 修改库存 */

// AdjustStock 手动调整库存
// 接收值：
//
//	userID - 操作人ID
//	req - 调整信息（包含调整目标，调整量和调整原因）
//
// 返回值：
//
//	*response.InventoryAdjustResp - 调整后响应信息（调整前后库存变化）
//	error - 错误信息
//
// 已迁 product-service:实现是"行锁 + 写流水 + 改库存"三步同事务,
// 这两张表都在 product-service 的库里,本地事务够不着。
func (is *InventoryService) AdjustStock(userID int64, req requset.InventoryAdjustReq) (*response.InventoryAdjustResp, error) {
	resp, errMsg, err := is.ProductRPC.AdjustStock(req.SkuId, req.ChangeQty, req.Remark, userID)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return resp, nil
}
