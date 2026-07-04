package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductRepo 商品表数据层实例
type ProductRepo struct {
	DB *gorm.DB
}

// NewProductRepo 新建商品表数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*ProductRepo - 商品表数据层实例指针
func NewProductRepo() *ProductRepo {
	return &ProductRepo{
		DB: db.DB,
	}
}

// WithTx 商品表事务实例
// 接收值：db - 数据库事务实例
// 返回值：*ProductRepo - 绑定事务的商品表数据层指针
func (p *ProductRepo) WithTx(db *gorm.DB) *ProductRepo {
	return &ProductRepo{
		DB: db,
	}
}

// CreateSpu 创建商品SPU
// 接收值：spu - SPU对象指针
// 返回值：
//
//	int64 - 新建SPU的唯一标识
//	error - 错误信息
func (p *ProductRepo) CreateSpu(spu *model.SysProductSpu) (spuId int64, err error) {
	err = p.DB.Create(&spu).Error
	spuId = spu.SpuId
	return spuId, err
}

// CreateSku 创建SKU表
// 接收值：sku - sku表对象指针
// 返回值：error - 错误信息
func (p *ProductRepo) CreateSku(sku *model.SysProductSku) error {
	return p.DB.Create(&sku).Error
}

// CreateImage 创建商品图片
// 接收值：image - 图片对象指针
// 返回值：error - 错误信息
func (p *ProductRepo) CreateImage(image *model.SysProductSpuImage) error {
	return p.DB.Create(&image).Error
}

// CheckSpecValue 检查同一 SPU 下是否已存在完全相同的规格组合
// 接收值：
//
//	spuId - spu唯一标识
//	specValues - sku规格值
//
// 返回值：
//
//	*model.SysProductSku - 所查询对应规格的sku信息
//	error - 错误信息
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

// GetSpuById 查询SPU信息(按spuID查)
// 接收值：spuId - 所查询SPU唯一标识
// 返回值：
//
//	*model.SysProductSpu - 所查询spu信息
//	error - 错误信息
func (p *ProductRepo) GetSpuById(spuId int64) (*model.SysProductSpu, error) {
	var spu model.SysProductSpu
	if err := p.DB.Where("spu_id = ?", spuId).First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &spu, nil
}

// GetSkuBySkuCode 查询SKU信息(按SKU编码查)
// 接收值：skuCode - 所查询SKU编码
// 返回值：
//
//	*model.SysProductSku - 所查询SKU信息
//	error - 错误信息
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

// GetSpuByCategory 查询SPU信息(按类目ID查)
// 接收值：categoryId - 所查询类目唯一标识
// 返回值：
//
//	*model.SysProductSpu - 所查询SPU信息
//	error - 错误信息
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

// GetProductStatus 查询商品状态(按SPU ID查)
// 接收值：spuId - 所查询SPU唯一标识
// 返回值：
//
//	string - 商品状态(draft/published/withdrawn)
//	error - 错误信息
func (p *ProductRepo) GetProductStatus(spuId int64) (string, error) {
	var spu model.SysProductSpu
	if err := p.DB.Where("spu_id = ?", spuId).First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return spu.SpuStatus, nil

}

// GetSpuList 分页查询商品SPU列表(含SKU聚合数据：最低/最高售价、总库存、总销量)
// 接收值：req - 商品查询请求(含分页、筛选、排序参数)
// 返回值：
//
//	[]model.SpuWithAgg - 商品SPU聚合列表
//	int64 - 总条数
//	error - 错误信息
func (p *ProductRepo) GetSpuList(req requset.SpuQueryReq) ([]model.SpuWithAgg, int64, error) {
	spuTable := model.SysProductSpu{}.TableName()
	skuTable := model.SysProductSku{}.TableName()

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
		Select(spuTable+`.*,
			COALESCE(MIN(`+skuTable+`.price), 0)      AS min_price,
			COALESCE(MAX(`+skuTable+`.price), 0)      AS max_price,
			COALESCE(SUM(`+skuTable+`.stock), 0)      AS total_stock,
			COALESCE(SUM(`+skuTable+`.sold_count), 0)  AS total_sold`).
		Joins("LEFT JOIN "+skuTable+" ON "+spuTable+".spu_id = "+skuTable+".spu_id AND "+skuTable+".is_deleted = ? AND "+skuTable+".sku_status = ?", false, "active").
		Group(spuTable + ".spu_id").
		Order(orderClause).
		Offset(offset).Limit(req.PageSize).
		Find(&rows).Error

	return rows, total, err
}

// GetSku 查询SKU信息(按SKU ID查)
// 接收值：skuId - 所查询SKU唯一标识
// 返回值：
//
//	*model.SysProductSku - 所查询SKU信息
//	error - 错误信息
func (p *ProductRepo) GetSku(skuId int64) (*model.SysProductSku, error) {
	var sku model.SysProductSku
	if err := p.DB.Where("sku_id = ?", skuId).First(&sku).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}

// GetSkuListBySpuId 查询SKU列表(按SPU ID查，排除已删除)
// 接收值：spuId - 所属SPU唯一标识
// 返回值：
//
//	[]model.SysProductSku - 该SPU下未删除的SKU列表
//	error - 错误信息
func (p *ProductRepo) GetSkuListBySpuId(spuId int64) ([]model.SysProductSku, error) {
	var skuList []model.SysProductSku
	if err := p.DB.Where("spu_id = ? AND is_deleted = ?", spuId, false).Find(&skuList).Error; err != nil {
		return nil, err
	}
	return skuList, nil
}

// GetImage 查询图片信息(按图片ID查)
// 接收值：imageId - 所查询图片唯一标识
// 返回值：
//
//	*model.SysProductSpuImage - 所查询图片信息
//	error - 错误信息
func (p *ProductRepo) GetImage(imageId int64) (*model.SysProductSpuImage, error) {
	var image model.SysProductSpuImage
	if err := p.DB.Where("image_id = ?", imageId).First(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageListBySpuId 查询图片列表(按SPU ID查)
// 接收值：spuId - 所属SPU唯一标识
// 返回值：
//
//	[]*model.SysProductSpuImage - 该SPU下的图片列表
//	error - 错误信息
func (p *ProductRepo) GetImageListBySpuId(spuId int64) ([]*model.SysProductSpuImage, error) {
	var imageList []*model.SysProductSpuImage
	if err := p.DB.Where("spu_id = ?", spuId).Find(&imageList).Error; err != nil {
		return nil, err
	}
	return imageList, nil
}

// UpdateSpu 更新商品SPU信息
// 接收值：spu - 待更新的SPU对象指针
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSpu(spu *model.SysProductSpu) error {
	return p.DB.Save(spu).Error
}

// UpdateSku 更新商品SKU信息
// 接收值：sku - 待更新的SKU对象指针
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateSku(sku *model.SysProductSku) error {
	return p.DB.Save(sku).Error
}

// UpdateImage 更新商品图片信息
// 接收值：image - 待更新的图片对象指针
// 返回值：error - 错误信息
func (p *ProductRepo) UpdateImage(image *model.SysProductSpuImage) error {
	return p.DB.Save(image).Error
}

// DeleteSpu 硬删除指定SPU(按SPU ID)
// 接收值：spuId - 待删除SPU唯一标识
// 返回值：error - 错误信息
func (p *ProductRepo) DeleteSpu(spuId int64) error {
	return p.DB.Where("spu_id = ?", spuId).Delete(&model.SysProductSpu{}).Error
}

// DeleteSku 硬删除指定SPU下所有SKU(按SPU ID)
// 接收值：spuId - 所属SPU唯一标识
// 返回值：error - 错误信息
func (p *ProductRepo) DeleteSku(spuId int64) error {
	return p.DB.Where("spu_id = ?", spuId).Delete(&model.SysProductSku{}).Error
}

// DeleteSpuById 软删除商品SPU(设置is_deleted=true)
// 接收值：spuId - 待删除SPU唯一标识
// 返回值：error - 错误信息
func (p *ProductRepo) DeleteSpuById(spuId int64) error {
	return p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ?", spuId, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(), // 同步更新修改时间，保留审计痕迹
		}).Error
}

// BatchDeleteSkuById 批量软删除指定SPU下所有未删除SKU(设置is_deleted=true)
// 接收值：spuId - 所属SPU唯一标识
// 返回值：error - 错误信息
func (p *ProductRepo) BatchDeleteSkuById(spuId int64) error {
	return p.DB.Model(&model.SysProductSku{}).
		Where("spu_id = ? AND is_deleted = ?", spuId, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}

// PublishProduct 上架商品(将draft/withdrawn状态变更为published)
// 接收值：spuId - 待上架SPU唯一标识
// 返回值：error - 错误信息(商品不存在或状态不合法时返回gorm.ErrRecordNotFound)
func (p *ProductRepo) PublishProduct(spuId int64) error {
	result := p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ? AND spu_status IN (?)", spuId, false, []string{"draft", "withdrawn"}).
		Update("spu_status", "published")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ValidateSpuPublish 校验商品是否满足上架条件
// 检查项：至少有一个启用(active)SKU、库存>0、价格>0
// 接收值：spuId - 待校验SPU唯一标识
// 返回值：error - 不满足条件时返回具体业务错误，满足时返回nil
func (p *ProductRepo) ValidateSpuPublish(spuId int64) error {
	var res model.SkuValidateResult

	err := p.DB.Model(&model.SysProductSku{}).
		Where("spu_id = ? AND is_deleted = ? AND sku_status = ?", spuId, false, "active").
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

// SoftDeleteSkusExcept 软删除指定SPU下不在保留列表中的SKU(设置is_deleted=true)
// keepIds为空时删除该SPU下全部未删除的SKU
// 接收值：
//
//	spuId - 所属SPU唯一标识
//	keepIds - 需保留的SKU ID列表
//
// 返回值：error - 错误信息
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

// DeleteImageByIds 硬删除指定ID的图片
// 接收值：ids - 待删除图片ID列表
// 返回值：error - 错误信息
func (p *ProductRepo) DeleteImageByIds(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return p.DB.Where("image_id IN ?", ids).Delete(&model.SysProductSpuImage{}).Error
}

// WithdrawProduct 下架商品(将published状态变更为withdrawn)
// 接收值：spuId - 待下架SPU唯一标识
// 返回值：error - 错误信息(商品不存在或状态不合法时返回gorm.ErrRecordNotFound)
func (p *ProductRepo) WithdrawProduct(spuId int64) error {
	result := p.DB.Model(&model.SysProductSpu{}).
		Where("spu_id = ? AND is_deleted = ? AND spu_status = ?", spuId, false, "published").
		Update("spu_status", "withdrawn")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
