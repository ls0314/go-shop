package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ============================================================
//	库存操作部分
// ============================================================

/* 查询低库存*/

// GetWarnStock  查询符合状态的低于阈值的库存信息列表
// 接收值 ： req - 查询参数（状态和阈值）
// 返回值：
//
//	[]response.InventoryWarnResp - 符合条件的库存信息列表
//	error - 错误信息
func (p *ProductRepo) GetWarnStock(req requset.InventoryWarnReq) ([]response.InventoryWarnResp, error) {
	spuTable := model.SysProductSpu{}.TableName()
	skuTable := model.SysProductSku{}.TableName()
	var stockWarnResp []response.InventoryWarnResp

	err := p.DB.Table(skuTable).
		Select(skuTable+".sku_id, "+spuTable+".spu_name, "+skuTable+".sku_name, "+
			skuTable+".stock, "+skuTable+".lock_stock, "+skuTable+".sold_count, "+
			skuTable+".sku_status").
		Joins("LEFT JOIN "+spuTable+" ON "+spuTable+".spu_id = "+skuTable+".spu_id AND "+spuTable+".is_deleted = ?", false).
		Where(skuTable+".stock <= ? AND "+skuTable+".is_deleted = ?", req.Threshold, false).
		Where(spuTable+".spu_status = ?", req.SpuStatus).
		Order(skuTable + ".stock ASC").
		Find(&stockWarnResp).Error

	return stockWarnResp, err

}

/* 更新库存 */

// UpdateStock  更新库存信息
// 接收值 ：
//
//	skuId - 待更新sku唯一标识
//	stockNumber - 更新后的库存数量
//
// 返回值：
//
//	[]response.InventoryWarnResp - 符合条件的库存信息列表
//	error - 错误信息
func (p *ProductRepo) UpdateStock(skuId, stockNumber int64) error {
	err := p.DB.Model(&model.SysProductSku{}).Where("sku_id = ? AND is_deleted = ?", skuId, false).
		Update("stock", stockNumber).Error
	return err
}

// UpdateSkuStockForLock 锁定库存：stock减、lock_stock加（原子操作，须在事务内调用）
// 接收值：skuId - SKU ID, qty - 锁定数量
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSkuStockForLock(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock - ?", qty),
			"lock_stock": gorm.Expr("lock_stock + ?", qty),
		}).Error
}

// UpdateSkuStockForPay 支付减扣库存：lock_stock减、sold_count加（原子操作，须在事务内调用）
// 接收值：skuId - SKU ID, qty - 锁定数量
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSkuStockForPay(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"lock_stock": gorm.Expr("lock_stock - ?", qty),
			"sold_count": gorm.Expr("sold_count + ?", qty),
		}).Error
}

// UpdateSkuStockForRelease 取消订单释放库存：lock_stock减、stock加（原子操作，须在事务内调用）
// 接收值：skuId - SKU ID, qty - 锁定数量
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSkuStockForRelease(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", qty),
			"lock_stock": gorm.Expr("lock_stock - ?", qty),
		}).Error
}

// UpdateSkuStockForRefund 退货补充库存：sold_count减、stock加（原子操作，须在事务内调用）
// 接收值：skuId - SKU ID, qty - 锁定数量
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSkuStockForRefund(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", qty),
			"sold_count": gorm.Expr("sold_count - ?", qty),
		}).Error
}

// ============================================================
//	库存日志部分
// ============================================================

// InventoryLogRepo 库存日志表数据层实例
type InventoryLogRepo struct {
	db *gorm.DB
}

// NewInventoryRepo 新建库存日志表数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*InventoryLogRepo - 库存日志表数据层实例指针
func NewInventoryRepo() *InventoryLogRepo {
	return &InventoryLogRepo{
		db: db.DB,
	}
}

// WithTx 库存日志表事务实例
// 接收值：db - 数据库事务实例
// 返回值：*InventoryLogRepo - 绑定事务的库存日志表据层指针
func (il *InventoryLogRepo) WithTx(tx *gorm.DB) *InventoryLogRepo {
	return &InventoryLogRepo{
		db: tx,
	}
}

// ============================================================
// 创建库存变更日志
// ============================================================

// CreateInventoryLog 创建库存更改日志
// 接收值：inventoryLog - 库存日志对象指针
// 返回值：error - 错误信息
func (il *InventoryLogRepo) CreateInventoryLog(inventoryLog *model.SysProductStockLog) (int64, error) {
	err := il.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&inventoryLog)
	return err.RowsAffected, err.Error
}

// CheckOrderLogExists 幂等检查：同一订单+变更类型是否已存在日志
// 接收值：orderId - 订单ID, changeType - 变更类型
// 返回值：bool - 是否已存在, error - 错误信息
func (il *InventoryLogRepo) CheckOrderLogExists(skuId, orderId int64, changeType string) (bool, error) {
	var count int64
	err := il.db.Model(&model.SysProductStockLog{}).
		Where("sku_id = ? AND order_id = ? AND change_type = ?", skuId, orderId, changeType).
		Count(&count).Error
	return count > 0, err
}

// ============================================================
// 查询库存变更日志
// ============================================================

// GetStockLogList 分页查询库存更改日志列表（联表返回spu_name)
// 接收值：req - 库存日志查询请求参数（含分页，筛选，时间段参数）
// 返回值：
//
//	[]response.InventoryLogList - 库存变更日志列表
//	int64 - 总条数
//	error - 错误信息
func (il *InventoryLogRepo) GetStockLogList(req requset.InventoryLogReq) ([]response.InventoryLogList, int64, error) {
	stockTable := model.SysProductStockLog{}.TableName()
	skuTable := model.SysProductSku{}.TableName()
	spuTable := model.SysProductSpu{}.TableName()

	// 联表处理（日志表、spu表、sku表）
	baseQuery := il.db.Table(stockTable).
		Select(stockTable + ".*, " + skuTable + ".sku_name, " + spuTable + ".spu_name").
		Joins("LEFT JOIN " + skuTable + " ON " + skuTable + ".sku_id = " + stockTable + ".sku_id").
		Joins("LEFT JOIN " + spuTable + " ON " + spuTable + ".spu_id = " + skuTable + ".spu_id")

	// 按参数过滤库存更改日志
	if req.SpuId != nil {
		baseQuery = baseQuery.Where(skuTable+".spu_id = ?", *req.SpuId)
	}

	if req.SkuId != nil {
		baseQuery = baseQuery.Where(stockTable+".sku_id = ?", *req.SkuId)
	}

	if req.ChangeType != "" {
		baseQuery = baseQuery.Where(stockTable+".change_type = ?", req.ChangeType)
	}

	if req.StartTime != nil {
		baseQuery = baseQuery.Where(stockTable+".created_at >= ?", req.StartTime)
	}

	if req.EndTime != nil {
		endOfDay := req.EndTime.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
		baseQuery = baseQuery.Where(stockTable+".created_at <= ?", endOfDay)
	}

	// 获取总条数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize

	var list []response.InventoryLogList

	// 聚合查询，按创建时间降序，日志ID降序排列
	err := baseQuery.Order(stockTable + ".created_at DESC, " + stockTable + ".log_id DESC").Offset(offset).Limit(pageSize).Find(&list).Error

	// 返回查询结果
	return list, total, err
}
