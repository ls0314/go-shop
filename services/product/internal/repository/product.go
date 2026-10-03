package repository

import (
	"demo-shop/services/product/internal/model"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductRepo 商品表数据层实例
type ProductRepo struct {
	DB *gorm.DB
}

const (
	spuTable      = "sys_product_spu"
	skuTable      = "sys_product_sku"
	categoryTable = "sys_category"

	// skuAggSelect 引用 skuTable,故必须是 var(常量不能引用变量)
	skuAggSelect = `COALESCE(MIN(` + skuTable + `.price), 0) AS min_price,
		COALESCE(MAX(` + skuTable + `.price), 0) AS max_price,
		COALESCE(SUM(` + skuTable + `.stock), 0) AS total_stock,
		COALESCE(SUM(` + skuTable + `.sold_count), 0) AS total_sold`

	skuJoinClause = "LEFT JOIN " + skuTable + " ON " + spuTable + ".spu_id = " + skuTable + ".spu_id AND " + skuTable + ".is_deleted = ? AND " + skuTable + ".sku_status = ?"
)

func NewProductRepo(conn *gorm.DB) *ProductRepo {
	return &ProductRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (p *ProductRepo) WithTx(tx *gorm.DB) *ProductRepo {
	return &ProductRepo{DB: tx}
}

// ============================================================
//	检查函数
// ============================================================

// CheckSpecValue 检查同一 SPU 下是否已存在完全相同的规格组合
func (p *ProductRepo) CheckSpecValue(spuId int64, specValues datatypes.JSONMap) (*model.SysProductSku, error) {
	var sku model.SysProductSku
	if err := p.DB.Where("spu_id = ? AND spec_values = ? AND is_deleted = ?", spuId, specValues, false).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

// ValidateSpuPublish 校验商品是否满足上架条件
// 检查项：至少有一个启用(active)SKU、库存>0、价格>0
func (p *ProductRepo) ValidateSpuPublish(spuId int64) error {
	var res model.SkuValidateResult

	err := p.DB.Model(&model.SysProductSku{}).
		Where("spu_id = ? AND is_deleted = ? AND sku_status = ?", spuId, false, model.SkuStatusActive).
		Select(`
					COUNT(*) AS sku_count,
					COALESCE(MAX(stock), 0) AS max_stock,
					COALESCE(MIN(price), 0) AS min_price
					`).
		Scan(&res).Error
	if err != nil {
		return err
	}

	if res.SkuCount == 0 {
		return model.ErrNoActiveSku
	}

	if res.MaxStock <= 0 {
		return model.ErrNoAvailableStock
	}

	if res.MinPrice <= 0 {
		return model.ErrInvalidPrice
	}
	return nil
}

// ============================================================
//	定义及实例化部分
// ============================================================

// CreateSpu 创建商品SPU,回填自增主键
func (p *ProductRepo) CreateSpu(spu *model.SysProductSpu) (spuId int64, err error) {
	err = p.DB.Create(&spu).Error
	spuId = spu.SpuId
	return spuId, err
}

// CreateSku 创建SKU
func (p *ProductRepo) CreateSku(sku *model.SysProductSku) error {
	return p.DB.Create(&sku).Error
}

// CreateImage 创建商品图片
func (p *ProductRepo) CreateImage(image *model.SysProductSpuImage) error {
	return p.DB.Create(&image).Error
}

// ============================================================
//	查询商品相关信息
// ============================================================

// GetSpuById 查询SPU信息(按spuID查)。
// 查不到直接返回 gorm.ErrRecordNotFound —— 调用方须用 errors.Is 判定,
// 不可当作基础设施故障(Marshal 后即上游的 error_msg 分支)。
func (p *ProductRepo) GetSpuById(spuId int64) (*model.SysProductSpu, error) {
	var spu model.SysProductSpu
	if err := p.DB.Where("spu_id = ?", spuId).First(&spu).Error; err != nil {
		return nil, err
	}
	return &spu, nil
}

// GetSkuBySkuCode 查询SKU信息(按SKU编码查)。查不到返回 (nil, nil)。
func (p *ProductRepo) GetSkuBySkuCode(skuCode string) (*model.SysProductSku, error) {
	var sku model.SysProductSku
	if err := p.DB.Where("sku_code = ?", skuCode).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

// GetSpuByCategory 查询SPU信息(按类目ID查)。查不到返回 (nil, nil)。
func (p *ProductRepo) GetSpuByCategory(categoryId int64) (*model.SysProductSpu, error) {
	var spu model.SysProductSpu
	if err := p.DB.Where("category_id = ?", categoryId).First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &spu, nil
}

// GetSpuList 分页查询商品SPU列表(含SKU聚合数据：最低/最高售价、总库存、总销量)
func (p *ProductRepo) GetSpuList(req model.SpuQueryReq) ([]model.SpuWithAgg, int64, error) {
	// WHERE 条件使用表前缀，避免 JOIN 后列名歧义
	baseQuery := p.DB.Table(spuTable).Where(spuTable+".is_deleted = ?", false)

	if req.SpuName != "" {
		baseQuery = baseQuery.Where(spuTable+".spu_name LIKE ?", "%"+req.SpuName+"%")
	}
	if req.CategoryId != nil {
		baseQuery = baseQuery.Where(spuTable+".category_id = ?", req.CategoryId)
	}
	if req.SpuStatus != "" {
		baseQuery = baseQuery.Where(spuTable+".spu_status = ?", req.SpuStatus)
	}
	if req.Brand != "" {
		baseQuery = baseQuery.Where(spuTable+".brand = ?", req.Brand)
	}

	// 单独 COUNT，不走 JOIN
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// JOIN 聚合查询，一次拿全 SPU + 价格/库存/销量
	var rows []model.SpuWithAgg
	offset := (req.Page - 1) * req.PageSize

	// 排序：Sort 不为空时使用用户端排序，否则使用管理端默认排序
	orderClause := spuTable + ".priority DESC, " + spuTable + ".updated_at DESC"
	switch req.Sort {
	case "sold":
		orderClause = "total_sold DESC"
	case "price_asc":
		orderClause = "min_price ASC"
	case "price_desc":
		orderClause = "max_price DESC"
	case "newest":
		orderClause = spuTable + ".created_at DESC"
	}

	err := baseQuery.
		Select(spuTable+".*",
			skuAggSelect).
		Joins(skuJoinClause, false, model.SkuStatusActive).
		Group(spuTable + ".spu_id").
		Order(orderClause).
		Offset(offset).Limit(req.PageSize).
		Find(&rows).Error

	return rows, total, err
}

// GetSku 查询SKU信息(按SKU ID查)
func (p *ProductRepo) GetSku(skuId int64) (*model.SysProductSku, error) {
	var sku model.SysProductSku
	if err := p.DB.Where("sku_id = ?", skuId).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrSkuNotExist
		}
		return nil, err
	}
	return &sku, nil
}

// GetSkuForUpdate 查询SKU并加行锁(用于库存并发操作)
func (p *ProductRepo) GetSkuForUpdate(skuId int64) (*model.SysProductSku, error) {
	var sku model.SysProductSku
	if err := p.DB.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sku_id = ?", skuId).
		First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrSkuNotExist
		}
		return nil, err
	}
	return &sku, nil
}

// GetSkuListBySpuId 查询SKU列表(按SPU ID查，排除已删除)
func (p *ProductRepo) GetSkuListBySpuId(spuId int64) ([]model.SysProductSku, error) {
	var skuList []model.SysProductSku
	if err := p.DB.Where("spu_id = ? AND is_deleted = ?", spuId, false).Find(&skuList).Error; err != nil {
		return nil, err
	}
	return skuList, nil
}

// GetSkuListByIds 批量查询 SKU(按 SKU ID 集合)。
//
// 不过滤 is_deleted:调用方(购物车校验)需要看到"已软删"这一事实才能给出
// "商品已下架"的提示;若在此过滤掉,调用方只能看到"ID 不存在",无法区分。
func (p *ProductRepo) GetSkuListByIds(skuIds []int64) ([]model.SysProductSku, error) {
	var skuList []model.SysProductSku
	if len(skuIds) == 0 {
		return skuList, nil
	}
	err := p.DB.Where("sku_id IN ?", skuIds).Find(&skuList).Error
	return skuList, err
}

// GetSpuListByIds 批量查询 SPU(按 SPU ID 集合),批量接口避免 N+1
func (p *ProductRepo) GetSpuListByIds(spuIds []int64) ([]model.SysProductSpu, error) {
	var spuList []model.SysProductSpu
	if len(spuIds) == 0 {
		return spuList, nil
	}
	err := p.DB.Where("spu_id IN ?", spuIds).Find(&spuList).Error
	return spuList, err
}

// GetImage 查询图片信息(按图片ID查)
func (p *ProductRepo) GetImage(imageId int64) (*model.SysProductSpuImage, error) {
	var image model.SysProductSpuImage
	if err := p.DB.Where("image_id = ?", imageId).First(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageListBySpuId 查询图片列表(按SPU ID查)
func (p *ProductRepo) GetImageListBySpuId(spuId int64) ([]*model.SysProductSpuImage, error) {
	var imageList []*model.SysProductSpuImage
	if err := p.DB.Where("spu_id = ?", spuId).Find(&imageList).Error; err != nil {
		return nil, err
	}
	return imageList, nil
}

// GetSpuESDoc 查询单个 SPU 的搜索文档
func (p *ProductRepo) GetSpuESDoc(spuId int64) (*model.SpuESDoc, error) {
	var doc model.SpuESDoc

	err := p.DB.Model(&model.SysProductSpu{}).
		Where(spuTable+".is_deleted = ? AND "+spuTable+".spu_id = ?", false, spuId).
		Select([]string{categoryTable + ".category_name",
			spuTable + ".*",
			skuAggSelect}).
		Joins(skuJoinClause, false, model.SkuStatusActive).
		Joins("LEFT JOIN " + categoryTable + " ON " + spuTable + ".category_id = " + categoryTable + ".category_id").
		Group(spuTable + ".spu_id, " + categoryTable + ".category_name").
		Scan(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetAllPublishedSpuEsDocs 查询全部已上架 SPU 的搜索文档(全量重建索引用)
func (p *ProductRepo) GetAllPublishedSpuEsDocs() ([]model.SpuESDoc, error) {
	var docs []model.SpuESDoc

	err := p.DB.Model(&model.SysProductSpu{}).
		Where(spuTable+".is_deleted = ? AND "+spuTable+".spu_status = ? ", false, model.SpuStatusPublished).
		Select([]string{categoryTable + ".category_name",
			spuTable + ".*",
			skuAggSelect}).
		Joins(skuJoinClause, false, model.SkuStatusActive).
		Joins("LEFT JOIN " + categoryTable + " ON " + spuTable + ".category_id = " + categoryTable + ".category_id").
		Group(spuTable + ".spu_id, " + categoryTable + ".category_name").
		Scan(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// GetChangedSpuEsDocs 查询自 since 起有变更的已上架 SPU 搜索文档(增量同步用)
func (p *ProductRepo) GetChangedSpuEsDocs(since time.Time) ([]model.SpuESDoc, error) {
	var docs []model.SpuESDoc

	err := p.DB.Model(&model.SysProductSpu{}).
		Where(spuTable+".is_deleted = ? AND "+spuTable+".spu_status = ?", false, model.SpuStatusPublished).
		Where("("+spuTable+".updated_at > ? OR EXISTS ("+
			"SELECT 1 FROM "+skuTable+" sku2 "+
			"WHERE sku2.spu_id = "+spuTable+".spu_id "+
			"AND sku2.is_deleted = false "+
			"AND sku2.updated_at > ?))", since, since).
		Select([]string{categoryTable + ".category_name",
			spuTable + ".*",
			skuAggSelect}).
		Joins(skuJoinClause, false, model.SkuStatusActive).
		Joins("LEFT JOIN " + categoryTable + " ON " + spuTable + ".category_id = " + categoryTable + ".category_id").
		Group(spuTable + ".spu_id, " + categoryTable + ".category_name").
		Scan(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// GetAllActiveSkuStock 闸门对账用:全部在售 SKU 的 (sku_id, stock)。
func (p *ProductRepo) GetAllActiveSkuStock() ([]model.SysProductSku, error) {
	var list []model.SysProductSku
	err := p.DB.Select("sku_id", "stock").
		Where("is_deleted = ? AND sku_status = ?", false, model.SkuStatusActive).
		Find(&list).Error
	return list, err
}

// GetWarnStock 查询低于阈值的可用库存列表(管理端库存预警)。
// 只取 published SPU 下的未删除 SKU,按可用库存升序 —— 最缺的排最前。
// 从单体 repository/inventory_repo.go 平移(该方法挂在 ProductRepo 上)。
func (p *ProductRepo) GetWarnStock(threshold int, spuStatus string) ([]model.StockWarnItem, error) {
	var rows []model.StockWarnItem

	err := p.DB.Table(skuTable).
		Select(skuTable+".sku_id, "+spuTable+".spu_name, "+skuTable+".sku_name, "+
			skuTable+".stock, "+skuTable+".lock_stock, "+skuTable+".sold_count, "+
			skuTable+".sku_status").
		Joins("LEFT JOIN "+spuTable+" ON "+spuTable+".spu_id = "+skuTable+".spu_id AND "+spuTable+".is_deleted = ?", false).
		Where(skuTable+".stock <= ? AND "+skuTable+".is_deleted = ?", threshold, false).
		Where(spuTable+".spu_status = ?", spuStatus).
		Order(skuTable + ".stock ASC").
		Find(&rows).Error

	return rows, err
}

// ============================================================
// 更新商品相关信息
// ============================================================

// ============================================================
// 库存增减（B1 库存四操作使用，原子表达式须在事务内调用）
// ============================================================

// UpdateSkuStockForLock 锁定库存：stock减、lock_stock加
func (p *ProductRepo) UpdateSkuStockForLock(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock - ?", qty),
			"lock_stock": gorm.Expr("lock_stock + ?", qty),
		}).Error
}

// UpdateSkuStockForPay 支付减扣库存：lock_stock减、sold_count加
func (p *ProductRepo) UpdateSkuStockForPay(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"lock_stock": gorm.Expr("lock_stock - ?", qty),
			"sold_count": gorm.Expr("sold_count + ?", qty),
		}).Error
}

// UpdateSkuStockForRelease 取消订单释放库存：lock_stock减、stock加
func (p *ProductRepo) UpdateSkuStockForRelease(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", qty),
			"lock_stock": gorm.Expr("lock_stock - ?", qty),
		}).Error
}

// UpdateSkuStockForRefund 退货补充库存：sold_count减、stock加
func (p *ProductRepo) UpdateSkuStockForRefund(skuId, qty int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", qty),
			"sold_count": gorm.Expr("sold_count - ?", qty),
		}).Error
}

// UpdateSpu 更新商品SPU信息
func (p *ProductRepo) UpdateSpu(spu *model.SysProductSpu) error {
	return p.DB.Save(spu).Error
}

// UpdateSku 更新商品SKU信息
func (p *ProductRepo) UpdateSku(sku *model.SysProductSku) error {
	return p.DB.Save(sku).Error
}

// UpdateSkuSold 累加SKU销量
func (p *ProductRepo) UpdateSkuSold(skuId int64, quantity int64) error {
	return p.DB.Model(model.SysProductSku{}).
		Where("sku_id = ?", skuId).
		Updates(map[string]interface{}{
			"sold_count": gorm.Expr("sold_count + ?", quantity),
			"updated_at": time.Now(),
		}).Error
}

// UpdateStock 把库存设为指定值(手动调整用;非原子增量,调用方须在读时持行锁)
func (p *ProductRepo) UpdateStock(skuId, stockNumber int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("sku_id = ? AND is_deleted = ?", skuId, false).
		Update("stock", stockNumber).Error
}

// UpdateImage 更新商品图片信息
func (p *ProductRepo) UpdateImage(image *model.SysProductSpuImage) error {
	return p.DB.Save(image).Error
}

// ============================================================
// 删除商品相关信息
// ============================================================

// DeleteSpu 硬删除指定SPU(按SPU ID)
func (p *ProductRepo) DeleteSpu(spuId int64) error {
	return p.DB.Where("spu_id = ?", spuId).Delete(&model.SysProductSpu{}).Error
}

// DeleteSku 硬删除指定SPU下所有SKU(按SPU ID)
func (p *ProductRepo) DeleteSku(spuId int64) error {
	return p.DB.Where("spu_id = ?", spuId).Delete(&model.SysProductSku{}).Error
}

// DeleteSpuById 软删除商品SPU(设置is_deleted=true)
func (p *ProductRepo) DeleteSpuById(spuId int64) error {
	return p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ?", spuId, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(), // 同步更新修改时间，保留审计痕迹
		}).Error
}

// BatchDeleteSkuById 批量软删除指定SPU下所有未删除SKU(设置is_deleted=true)
func (p *ProductRepo) BatchDeleteSkuById(spuId int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("spu_id = ? AND is_deleted = ?", spuId, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}

// DeleteImageByIds 硬删除指定ID的图片
func (p *ProductRepo) DeleteImageByIds(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return p.DB.Where("image_id IN ?", ids).Delete(&model.SysProductSpuImage{}).Error
}

// SoftDeleteSkusExcept 软删除指定SPU下不在保留列表中的SKU(设置is_deleted=true)
// keepIds为空时删除该SPU下全部未删除的SKU
func (p *ProductRepo) SoftDeleteSkusExcept(spuId int64, keepIds []int64) error {
	query := p.DB.Model(&model.SysProductSku{}).
		Where("spu_id = ? AND is_deleted = ?", spuId, false)
	if len(keepIds) > 0 {
		query = query.Where("sku_id NOT IN ?", keepIds)
	}
	return query.Updates(map[string]interface{}{
		"is_deleted": true,
		"updated_at": time.Now(),
	}).Error
}

// ============================================================
// 商品上下架操作
// ============================================================

// PublishProduct 上架商品(将draft/withdrawn状态变更为published)
// 商品不存在或状态不合法时返回 gorm.ErrRecordNotFound
func (p *ProductRepo) PublishProduct(spuId int64) error {
	result := p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ? AND spu_status IN (?)", spuId, false,
			[]string{model.SpuStatusDraft, model.SpuStatusWithdrawn}).
		Update("spu_status", model.SpuStatusPublished)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// WithdrawProduct 下架商品(将published状态变更为withdrawn)
// 商品不存在或状态不合法时返回 gorm.ErrRecordNotFound
func (p *ProductRepo) WithdrawProduct(spuId int64) error {
	result := p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ? AND spu_status = ?", spuId, false, model.SpuStatusPublished).
		Update("spu_status", model.SpuStatusWithdrawn)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
