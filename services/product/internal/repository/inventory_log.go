package repository

import (
	"time"

	"demo-shop/services/product/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventoryLogRepo struct {
	DB *gorm.DB
}

func NewInventoryLogRepo(conn *gorm.DB) *InventoryLogRepo {
	return &InventoryLogRepo{DB: conn}
}

func (il *InventoryLogRepo) WithTx(tx *gorm.DB) *InventoryLogRepo {
	return &InventoryLogRepo{DB: tx}
}

// CreateInventoryLog 插入库存流水。返回受影响行数:0 表示
// (sku_id, order_id, change_type) 已存在,即本次操作已执行过。
func (il *InventoryLogRepo) CreateInventoryLog(inventoryLog *model.SysProductStockLog) (int64, error) {
	res := il.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(inventoryLog)
	return res.RowsAffected, res.Error
}

// StockLogQuery 库存流水分页查询条件。
// 各筛选项为指针/空值即表示不过滤;时间区间为闭区间(含当日)。
type StockLogQuery struct {
	Page     int
	PageSize int
	SkuId    *int64
	SpuId    *int64
	// OrderNo 按订单号过滤(追溯用)。原先是 OrderId,
	// 拆库后新流水不再写 order_id,只有 order_no 能覆盖全部记录
	OrderNo    string
	ChangeType string
	StartTime  *time.Time
	EndTime    *time.Time
}

// ListStockLogs 分页查询库存流水,联 sku / spu 取名称。
//
// 从单体 repository/inventory_repo.go 的 GetStockLogList 平移。
// 三张表(sock_log / sku / spu)都在本服务的库里,故这里是**同库联表**,
// 不存在跨库 JOIN —— 这也是它必须搬过来的原因:单体那边的写入已迁走,
// 留着只会读到一个停更的副本。
func (il *InventoryLogRepo) ListStockLogs(q StockLogQuery) ([]model.StockLogItem, int64, error) {
	stockTable := "sys_product_stock_log"
	skuTable := "sys_product_sku"
	spuTable := "sys_product_spu"

	baseQuery := il.DB.Table(stockTable).
		Select(stockTable + ".*, " + skuTable + ".sku_name, " +
			skuTable + ".spu_id, " + spuTable + ".spu_name").
		Joins("LEFT JOIN " + skuTable + " ON " + skuTable + ".sku_id = " + stockTable + ".sku_id").
		Joins("LEFT JOIN " + spuTable + " ON " + spuTable + ".spu_id = " + skuTable + ".spu_id")

	if q.SpuId != nil {
		baseQuery = baseQuery.Where(skuTable+".spu_id = ?", *q.SpuId)
	}
	if q.SkuId != nil {
		baseQuery = baseQuery.Where(stockTable+".sku_id = ?", *q.SkuId)
	}
	if q.ChangeType != "" {
		baseQuery = baseQuery.Where(stockTable+".change_type = ?", q.ChangeType)
	}
	if q.OrderNo != "" {
		baseQuery = baseQuery.Where(stockTable+".order_no = ?", q.OrderNo)
	}
	if q.StartTime != nil {
		baseQuery = baseQuery.Where(stockTable+".created_at >= ?", *q.StartTime)
	}
	if q.EndTime != nil {
		// 与原实现一致:结束日取当天 23:59:59,使"选同一天"能查到当天全部流水
		endOfDay := q.EndTime.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
		baseQuery = baseQuery.Where(stockTable+".created_at <= ?", endOfDay)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.PageSize
	var list []model.StockLogItem
	err := baseQuery.
		Order(stockTable + ".created_at DESC, " + stockTable + ".log_id DESC").
		Offset(offset).Limit(q.PageSize).
		Find(&list).Error

	return list, total, err
}
