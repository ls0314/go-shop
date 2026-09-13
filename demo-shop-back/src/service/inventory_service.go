package service

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"fmt"

	"gorm.io/gorm"
)

// ============================================================
//	定义及实例化
// ============================================================

// InventoryService 库存服务层实例
type InventoryService struct {
	ProductRepo      *repository.ProductRepo      // 商品数据层实例
	InventoryLogRepo *repository.InventoryLogRepo // 库存日志数据层实例
	db               *gorm.DB
}

// NewInventoryService 创建库存服务层实例
// 接收值：使用全局repository初始化，故无接收值
// 返回值：*InventoryService - 库存服务层实例指针
func NewInventoryService() *InventoryService {
	return &InventoryService{
		ProductRepo:      repository.NewProductRepo(),
		InventoryLogRepo: repository.NewInventoryRepo(),
		db:               db.DB,
	}
}

// ============================================================
//	库存操作部分(对外接口）
// ============================================================

/* 查询 */

// GetSkuStock 获取单个商品库存信息
// 接收值：skuId - 商品SKU ID
// 返回值：*response.SkuInventoryResp - 商品库存信息详情响应
func (is *InventoryService) GetSkuStock(skuId int64) (*response.SkuInventoryResp, error) {
	// 获取SKU和SPU信息
	sku, err := is.ProductRepo.GetSku(skuId)
	if err != nil {
		return nil, err
	}

	spu, err := is.ProductRepo.GetSpuById(sku.SpuId)
	if err != nil {
		return nil, err
	}

	// 过滤掉内部字段后拼接成响应信息
	stockResp := &response.SkuInventoryResp{
		SkuId:      sku.SkuId,
		SkuName:    sku.SkuName,
		SpuId:      sku.SpuId,
		SpuName:    spu.SpuName,
		SpecValues: sku.SpecValues,
		Stock:      sku.Stock,
		LockStock:  sku.LockStock,
		TotalStock: sku.Stock + sku.LockStock,
		SoldCount:  sku.SoldCount,
		SkuStatus:  sku.SkuStatus,
	}

	return stockResp, nil
}

// GetSkuStockListBySpu 获取spu下商品库存信息列表
// 接收值：spuId - 商品SPU ID
// 返回值：
//
//	*response.SkuInventoryListResp - 商品库存信息列表响应（聚合返回总库存、总销量等信息）
//	error - 错误信息
func (is *InventoryService) GetSkuStockListBySpu(spuId int64) (*response.SkuInventoryListResp, error) {

	// 查询spu信息及其下属的sku信息列表
	skuList, err := is.ProductRepo.GetSkuListBySpuId(spuId)
	if err != nil {
		return nil, err
	}

	spu, err := is.ProductRepo.GetSpuById(spuId)
	if err != nil {
		return nil, err
	}

	// 过滤内部字段后拼接成响应体
	listResp := make([]response.SkuInventoryResp, 0, len(skuList))
	var totalStock, totalLock, totalSold int64
	for _, sku := range skuList {
		// 聚合字段获得总库存等信息
		totalStock += sku.Stock
		totalLock += sku.LockStock
		totalSold += sku.SoldCount
		listResp = append(listResp, response.SkuInventoryResp{
			SkuId:      sku.SkuId,
			SkuName:    sku.SkuName,
			SpuId:      sku.SpuId,
			SpuName:    spu.SpuName,
			SpecValues: sku.SpecValues,
			Stock:      sku.Stock,
			LockStock:  sku.LockStock,
			TotalStock: sku.Stock + sku.LockStock,
			SoldCount:  sku.SoldCount,
			SkuStatus:  sku.SkuStatus,
		})
	}

	// 返回最终响应
	return &response.SkuInventoryListResp{
		List:       listResp,
		TotalStock: totalStock,
		TotalLock:  totalLock,
		TotalSold:  totalSold,
	}, nil
}

// GetWarnStockList 获取低于库存阈值的商品库存信息列表
// 接收值：req - 查询参数包含阈值和spu状态信息
// 返回值：
//
//	*response.SkuInventoryListResp - 低于阈值的商品库存信息列表响应（过滤去除内部字段）
//	error - 错误信息
func (is *InventoryService) GetWarnStockList(req requset.InventoryWarnReq) (*[]response.InventoryWarnResp, error) {
	// 若未传入查询参数则设置默认值
	if req.Threshold == 0 {
		req.Threshold = 10
	}
	if req.SpuStatus == "" {
		req.SpuStatus = model.SpuStatusPublished
	}
	// 调用数据层查询商品库存信息列表
	stockWarnList, err := is.ProductRepo.GetWarnStock(req)
	if err != nil {
		return nil, err
	}

	return &stockWarnList, err
}

// GetStockLogList 分页获取库存变更日志列表
// 接收值： req - 查询参数（包含分页、筛选条件等信息）
//
// 返回值：
//
//	*response.InventoryLogResp - 库存变更日志列表
//	error - 错误信息
func (is *InventoryService) GetStockLogList(req requset.InventoryLogReq) (*response.InventoryLogResp, error) {
	// 保证传入页面信息合法性
	if req.Page <= 0 {
		req.Page = 1
	}
	// 防参数越界:<=0 用默认 50;>50 封顶 50(而非压成 50,避免大 pageSize 反而返回最少)
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.PageSize > 50 {
		req.PageSize = 50
	}

	// 调用数据获取库存变更日志
	stockList, total, err := is.InventoryLogRepo.GetStockLogList(req)
	if err != nil {
		return nil, err
	}

	// 拼接响应体
	stockListResp := &response.InventoryLogResp{
		List:     stockList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return stockListResp, err
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
func (is *InventoryService) AdjustStock(userID int64, req requset.InventoryAdjustReq) (*response.InventoryAdjustResp, error) {
	var stockResp response.InventoryAdjustResp

	// 调整原因不能为空
	if req.Remark == "" {
		return nil, model.ErrRemarkEmpty
	}

	// 开启事务，保证一致性
	err := is.db.Transaction(func(tx *gorm.DB) error {
		// 开启事务实例
		productTx := is.ProductRepo.WithTx(tx)
		logTx := is.InventoryLogRepo.WithTx(tx)

		// 并发安全查询sku信息（当commit后才能对sku进行更新或删除）
		sku, err := productTx.GetSkuForUpdate(req.SkuId)
		if err != nil {
			return err
		}

		// 确保库存变动合法
		if sku.Stock+req.ChangeQty < 0 {
			return model.ErrStockNegative
		}

		// 构建响应信息
		stockResp = response.InventoryAdjustResp{
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + req.ChangeQty,
		}

		// 调用数据层用调整信息更新SKU库存信息
		err = productTx.UpdateStock(req.SkuId, sku.Stock+req.ChangeQty)
		if err != nil {
			return err
		}

		// 新建库存变动日志，写入对应变动信息
		err = logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       sku.SkuId,
			ChangeType:  model.StockManualAdjust,
			ChangeQty:   req.ChangeQty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + req.ChangeQty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock,
			Remark:      req.Remark,
			CreateBy:    userID,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("sku:stock:%d", req.SkuId))
	}
	return &stockResp, nil
}

// ============================================================
//	库存操作部分(内部接口）
// ============================================================

// LockStock 下单锁定库存（内部接口，独立事务）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 锁定数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) LockStock(skuId, qty, orderId int64) error {
	return is.LockStockWithTx(is.db, skuId, qty, orderId)
}

// LockStockWithTx 下单锁定库存（共享外部事务）
// 接收值:
//
//	tx     - 外部事务（为 nil 时自动开启独立事务）
//	skuId  - SKU ID
//	qty    - 锁定数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) lockStockWithTx(tx *gorm.DB, skuId, qty, orderId int64) error {
	if tx == nil {
		tx = is.db
	}
	err := tx.Transaction(func(innerTx *gorm.DB) error {
		productTx := is.ProductRepo.WithTx(innerTx)
		logTx := is.InventoryLogRepo.WithTx(innerTx)

		// 幂等检查: 同一订单+同一操作类型已执行过则直接返回成功
		idempotent, err := logTx.CheckOrderLogExists(skuId, orderId, model.StockOrderLock)
		if err != nil {
			return err
		}
		if idempotent {
			return nil
		}

		// SELECT ... FOR UPDATE 锁定SKU行
		sku, err := productTx.GetSkuForUpdate(skuId)
		if err != nil {
			return err
		}

		// 校验SKU状态
		if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
			return model.ErrSkuDisabled
		}

		// 校验可用库存
		if sku.Stock < qty {
			return model.ErrStockNotEnough
		}

		// UPDATE stock = stock - qty, lock_stock = lock_stock + qty
		if err := productTx.UpdateSkuStockForLock(skuId, qty); err != nil {
			return err
		}

		// 写入库存变更日志
		return logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       skuId,
			ChangeType:  model.StockOrderLock,
			ChangeQty:   qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock - qty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock + qty,
			OrderId:     orderId,
		})
	})
	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("sku:stock:%d", skuId))
	}
	return err
}

// DeductStock 支付减少锁定库存（内部接口，独立事务）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 减少数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) DeductStock(skuId, qty, orderId int64) error {
	return is.DeductStockWithTx(is.db, skuId, qty, orderId)
}

// DeductStockWithTx 支付减少锁定库存（共享外部事务）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 减少数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) DeductStockWithTx(tx *gorm.DB, skuId, qty, orderId int64) error {
	if tx == nil {
		tx = is.db
	}
	err := tx.Transaction(func(innerTx *gorm.DB) error {
		productTx := is.ProductRepo.WithTx(innerTx)
		logTx := is.InventoryLogRepo.WithTx(innerTx)

		// 幂等检查: 同一订单+同一操作类型已执行过则直接返回成功
		idempotent, err := logTx.CheckOrderLogExists(skuId, orderId, model.StockPayDeduct)
		if err != nil {
			return err
		}
		if idempotent {
			return nil
		}

		// SELECT ... FOR UPDATE 锁定SKU行
		sku, err := productTx.GetSkuForUpdate(skuId)
		if err != nil {
			return err
		}

		// 校验SKU状态
		if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
			return model.ErrSkuDisabled
		}

		// 校验可用库存
		if sku.LockStock < qty {
			return model.ErrStockNotEnough
		}

		// UPDATE lock_stock = lock_stock - qty, sold_count = sold_count + qty
		if err := productTx.UpdateSkuStockForPay(skuId, qty); err != nil {
			return err
		}

		// 写入库存变更日志
		return logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       skuId,
			ChangeType:  model.StockPayDeduct,
			ChangeQty:   qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock - qty,
			OrderId:     orderId,
		})

	})
	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("sku:stock:%d", skuId))
	}
	return err
}

// ReleaseStock 取消订单释放库存（内部接口，独立事务）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 锁定数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) ReleaseStock(skuId, qty, orderId int64) error {
	return is.ReleaseStockWithTx(is.db, skuId, qty, orderId)
}

// ReleaseStockWithTx 取消订单释放库存（共享外部事务）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 锁定数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) ReleaseStockWithTx(tx *gorm.DB, skuId, qty, orderId int64) error {
	if tx == nil {
		tx = is.db
	}
	err := is.db.Transaction(func(inner *gorm.DB) error {
		productTx := is.ProductRepo.WithTx(inner)
		logTx := is.InventoryLogRepo.WithTx(inner)

		// 幂等检查: 同一订单+同一操作类型已执行过则直接返回成功
		idempotent, err := logTx.CheckOrderLogExists(skuId, orderId, model.StockOrderRelease)
		if err != nil {
			return err
		}
		if idempotent {
			return nil
		}

		// SELECT ... FOR UPDATE 锁定SKU行
		sku, err := productTx.GetSkuForUpdate(skuId)
		if err != nil {
			return err
		}

		// 校验SKU状态
		if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
			return model.ErrSkuDisabled
		}

		// 校验可用库存
		if sku.LockStock < qty {
			return model.ErrStockNegative
		}

		// UPDATE stock = stock + qty, lock_stock = lock_stock - qty
		if err := productTx.UpdateSkuStockForRelease(skuId, qty); err != nil {
			return err
		}

		// 写入库存变更日志
		return logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       skuId,
			ChangeType:  model.StockOrderRelease,
			ChangeQty:   qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + qty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock - qty,
			OrderId:     orderId,
		})

	})
	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("sku:stock:%d", skuId))
	}
	return err
}

// RefundStock 退款增加库存（内部接口，供 OrderService 调用）
// 并发策略: SELECT ... FOR UPDATE 行锁
// 幂等保障: 同一 order_id 的 order_lock 操作仅执行一次
//
// 接收值:
//
//	skuId  - SKU ID
//	qty    - 锁定数量
//	orderId - 关联订单ID
//
// 返回值: error - 错误信息
func (is *InventoryService) RefundStock(skuId, qty, orderId int64) error {
	err := is.db.Transaction(func(tx *gorm.DB) error {
		productTx := is.ProductRepo.WithTx(tx)
		logTx := is.InventoryLogRepo.WithTx(tx)

		// 幂等检查: 同一订单+同一操作类型已执行过则直接返回成功
		idempotent, err := logTx.CheckOrderLogExists(skuId, orderId, model.StockRefundRelease)
		if err != nil {
			return err
		}
		if idempotent {
			return nil
		}

		// SELECT ... FOR UPDATE 锁定SKU行
		sku, err := productTx.GetSkuForUpdate(skuId)
		if err != nil {
			return err
		}

		// 校验SKU状态
		if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
			return model.ErrSkuDisabled
		}

		// UPDATE stock = stock + qty, sold_count = sold_count - qty
		if err := productTx.UpdateSkuStockForRefund(skuId, qty); err != nil {
			return err
		}

		// 写入库存变更日志
		return logTx.CreateInventoryLog(&model.SysProductStockLog{
			SkuId:       skuId,
			ChangeType:  model.StockRefundRelease,
			ChangeQty:   qty,
			BeforeStock: sku.Stock,
			AfterStock:  sku.Stock + qty,
			BeforeLock:  sku.LockStock,
			AfterLock:   sku.LockStock,
			OrderId:     orderId,
		})

	})
	// 库存变更后失效缓存(短 TTL 兜底,失效失败最多 30 秒旧值)
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("sku:stock:%d", skuId))
	}
	return err
}
