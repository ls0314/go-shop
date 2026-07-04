package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"encoding/json"
	"fmt"

	"errors"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// specItem 规格模板中的单条规格
type specItem struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// validateSpecSku 校验规格模板与SKU列表
// 接收值：spuId - 商品SPU ID（新建时为0）
//
//	specTemplate - 规格模板（JSON格式）
//	skuList - SKU列表
//	repo - 商品数据层实例
//	checkSpecValue - 是否校验规格组合唯一性
//
// 返回值：error - 错误信息
func validateSpecSku(spuId int64, specTemplate interface{}, skuList []model.SysProductSku, repo *repository.ProductRepo, checkSpecValue bool) error {
	// 将 specTemplate 转为 []specItem（兼容 JSONMap / JSON / []byte 等格式）
	b, err := json.Marshal(specTemplate)
	if err != nil {
		return model.ErrSpuTemplate
	}
	var specs []specItem
	if err := json.Unmarshal(b, &specs); err != nil || len(specs) == 0 {
		return model.ErrSpuTemplate
	}

	// 规则1：规格模板格式校验
	for _, s := range specs {
		if s.Name == "" || len(s.Values) == 0 {
			return model.ErrSpuTemplate
		}
	}

	// 规则2：SKU 数量 = 笛卡尔积
	expected := 1
	for _, s := range specs {
		expected *= len(s.Values)
	}
	if len(skuList) != expected {
		return model.ErrSkuNum
	}

	// 规则3+4：逐 SKU 校验
	seen := make(map[string]bool, len(skuList))
	for _, sku := range skuList {
		// 规则3：spec_values 的 key 和 value 必须与 specTemplate 匹配
		if len(sku.SpecValues) != len(specs) {
			return model.ErrSpecValues
		}
		for _, s := range specs {
			val, ok := sku.SpecValues[s.Name]
			if !ok {
				return model.ErrSpecValues
			}
			if !contains(s.Values, fmt.Sprint(val)) {
				return model.ErrSpecValues
			}
		}

		// 规则4：规格组合唯一（请求内去重 + DB 去重）
		if checkSpecValue == true {
			key := fmt.Sprint(sku.SpecValues)
			if seen[key] {
				return model.ErrSpecValuesNotOnly
			}
			seen[key] = true

			existing, err := repo.CheckSpecValue(spuId, sku.SpecValues)
			if err != nil {
				return err
			}
			if existing != nil && existing.SkuId != sku.SkuId {
				return model.ErrSpecValuesNotOnly
			}
		}
	}

	return nil
}

// contains 判断切片是否包含目标字符串
// 接收值：slice - 字符串切片
//
//	target - 目标字符串
//
// 返回值：bool - 是否包含
func contains(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// ProductService 商品服务层实例
type ProductService struct {
	ProductRepo  *repository.ProductRepo  // 商品数据层实例
	CategoryRepo *repository.CategoryRepo // 类目数据层实例
	db           *gorm.DB
}

// NewProductService 创建商品服务层实例
// 接收值：使用全局repository初始化，故无接收值
// 返回值：*ProductService - 商品服务层实例指针
func NewProductService() *ProductService {
	return &ProductService{
		ProductRepo:  repository.NewProductRepo(),
		CategoryRepo: repository.NewCategoryRepo(),
		db:           db.DB,
	}
}

// CreateProduct 创建商品
// 创建商品SPU，同时创建关联的SKU列表和图片列表，事务保证原子性
// 接收值：spu - 商品SPU对象指针（含SKU列表和图片列表）
// 返回值：int64 - 新建商品SPU ID
//
//	error - 错误信息
func (p *ProductService) CreateProduct(spu *model.SysProductSpu) (int64, error) {
	// 类目校验
	category, err := p.CategoryRepo.GetCategoryById(spu.CategoryId)
	if err != nil {
		return 0, err
	}
	if category == nil || category.Status != "active" || !category.IsLeaf {
		return 0, model.ErrCategoryNotUsed
	}

	// 规格与 SKU 校验
	if spu.SkuList == nil || len(*spu.SkuList) == 0 {
		return 0, model.ErrSkuListEmpty
	}
	if err := validateSpecSku(0, spu.SpecTemplate, *spu.SkuList, p.ProductRepo, true); err != nil {
		return 0, err
	}

	var respSpuId int64
	// 开启事务
	err = p.db.Transaction(func(tx *gorm.DB) error {
		txRepo := p.ProductRepo.WithTx(tx)

		// 调用数据层创建SPU
		spuId, err := txRepo.CreateSpu(spu)
		if err != nil {
			return err
		}
		// 取出spuId用于handler层响应
		respSpuId = spuId

		// 逐条创建sku表
		for _, sku := range *spu.SkuList {
			// 保持sku一致性
			sku := sku
			// 设置sku.SpuId 为刚才创建的spu对应spuId
			sku.SpuId = spuId
			// 校验sku信息，售价小于0返回错误
			if sku.Price <= 0 {
				return model.ErrInvalidPrice
			}
			// 校验skuCode的一致性
			if skuCode, err := txRepo.GetSkuBySkuCode(sku.SkuCode); err != nil || skuCode != nil {
				return model.ErrSkuCodeNotOnly
			}
			// 调用数据层创建sku信息
			if err := txRepo.CreateSku(&sku); err != nil {
				return err
			}
		}

		// 逐条创建图片表
		if spu.ImageList != nil {
			for _, image := range *spu.ImageList {
				// 保持图片一致性
				image := image
				// 设置image.SpuId 为刚才创建的spu对应spuId
				image.SpuId = spuId
				// 调用数据层创建图片信息
				if err := txRepo.CreateImage(&image); err != nil {
					return err
				}
			}
		}

		return nil
	})
	// 返回spuId和错误信息
	return respSpuId, err
}

// GetProductSpuList 管理端商品列表
// 分页查询商品列表，支持名称/类目/状态/品牌筛选，返回SKU聚合数据
// 接收值：req - 商品查询请求参数（含分页和筛选条件）
// 返回值：*response.GetProductListResp - 商品列表响应（含SPU信息、SKU聚合价格与库存）
//
//	error - 错误信息
func (p *ProductService) GetProductSpuList(req requset.SpuQueryReq) (*response.GetProductListResp, error) {
	// 保证传入页面信息合法性
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	// 调用数据层获取完整的spu分页列表信息
	rows, total, err := p.ProductRepo.GetSpuList(req)
	if err != nil {
		return nil, err
	}

	// 创建管理端响应体中的list字段
	list := make([]response.SpuList, 0, len(rows))
	for _, row := range rows {
		// 获取类目id对应的的类目信息
		category, err := p.CategoryRepo.GetCategoryById(row.CategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, model.CategoryNotExist
		}
		// 将完整spu结构筛选字段后写入管理端响应体list字段
		list = append(list, response.SpuList{
			SpuId:        row.SpuId,
			SpuName:      row.SpuName,
			CategoryId:   row.CategoryId,
			CategoryName: category.CategoryName,
			Brand:        row.Brand,
			MainImage:    row.MainImage,
			SpuStatus:    row.SpuStatus,
			Priority:     row.Priority,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			MinPrice:     int64(row.MinPrice),
			MaxPrice:     int64(row.MaxPrice),
			TotalStock:   row.TotalStock,
			TotalSold:    row.TotalSold,
		})
	}

	// 返回管理端响应体结构体
	return &response.GetProductListResp{
		List:     &list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UserGetProductSpuList 用户端商品列表
// 公开接口，返回已上架+未删除的商品列表，按销量和优先级排序
// 接收值：req - 商品查询请求参数（仅返回spu_status=published的商品）
// 返回值：*response.UserGetProductListResp - 用户端商品列表响应（不含成本价等内部字段）
//
//	error - 错误信息
func (p *ProductService) UserGetProductSpuList(req requset.SpuQueryReq) (*response.UserGetProductListResp, error) {
	// 保证传入页面信息合法性
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 50 {
		req.PageSize = 10
	}

	// 用户端只展示已上架商品
	req.SpuStatus = "published"
	// 调用数据层获取完整的spu分页列表信息
	rows, total, err := p.ProductRepo.GetSpuList(req)
	if err != nil {
		return nil, err
	}
	// 创建用户端响应体中的list字段
	list := make([]response.UserSpuList, 0, len(rows))
	for _, row := range rows {
		// 获取类目id对应的的类目信息
		category, err := p.CategoryRepo.GetCategoryById(row.CategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, model.CategoryNotExist
		}
		// 将完整spu结构筛选字段后写入用户端响应体list字段
		list = append(list, response.UserSpuList{
			SpuId:        row.SpuId,
			SpuName:      row.SpuName,
			CategoryName: category.CategoryName,
			Brand:        row.Brand,
			MainImage:    row.MainImage,
			MinPrice:     int64(row.MinPrice),
			MaxPrice:     int64(row.MaxPrice),
			TotalSold:    row.TotalSold,
		})
	}
	// 返回用户端响应体结构体
	return &response.UserGetProductListResp{
		List:     &list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetProduct 管理端商品详情
// 获取商品完整信息，包含SKU列表（含成本价/锁定库存）和图片列表
// 接收值：spuId - 商品SPU ID
// 返回值：*response.GetProductResp - 商品详情响应（含SKU和图片完整信息）
//
//	error - 错误信息
func (p *ProductService) GetProduct(spuId int64) (*response.GetProductResp, error) {
	var spuResp response.GetProductResp
	var skuListResp []response.SkuList
	var imageListResp []response.ImageList

	// 调用数据层根据spuID获取spu信息
	spu, err := p.ProductRepo.GetSpuById(spuId)
	if err != nil {
		return nil, err
	}
	if spu == nil {
		return nil, nil
	}

	// 调用数据层获取spu对应的类目信息
	category, err := p.CategoryRepo.GetCategoryById(spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, model.CategoryNotExist
	}

	// 调用数据层获取spu下面全部的sku列表信息
	skuList, err := p.ProductRepo.GetSkuListBySpuId(spuId)
	if err != nil {
		return nil, err
	}

	// 调用数据层获取spu下面全部的图像列表信息
	imageList, err := p.ProductRepo.GetImageListBySpuId(spuId)
	if err != nil {
		return nil, err
	}

	// 根据sku列表信息 逐条构建管理端响应信息
	for _, sku := range skuList {
		skuResp := response.SkuList{
			SkuId:     sku.SkuId,
			SpuId:     sku.SpuId,
			SkuName:   sku.SkuName,
			SpecValue: sku.SpecValues,
			Price:     sku.Price,
			CostPrice: sku.CostPrice,
			Stock:     sku.Stock,
			LockStock: sku.LockStock,
			SoldCount: sku.SoldCount,
			SkuCode:   sku.SkuCode,
			SkuImage:  sku.SkuImage,
			SkuStatus: sku.SkuStatus,
		}
		skuListResp = append(skuListResp, skuResp)
	}

	// 根据图像列表信息 逐条构建管理端响应信息
	for _, image := range imageList {
		imageResp := response.ImageList{
			ImageId:   image.ImageId,
			ImageUrl:  image.ImageUrl,
			SortOrder: image.SortOrder,
			IsMain:    image.IsMain,
		}
		imageListResp = append(imageListResp, imageResp)
	}

	// 整合全部以获取信息，构建完整管理端响应体信息
	spuResp = response.GetProductResp{
		SpuId:        spu.SpuId,
		SpuName:      spu.SpuName,
		CategoryId:   category.CategoryId,
		CategoryName: category.CategoryName,
		Brand:        spu.Brand,
		Description:  spu.Description,
		MainImage:    spu.MainImage,
		SpecTemplate: spu.SpecTemplate,
		SpuStatus:    spu.SpuStatus,
		Priority:     spu.Priority,
		SkuList:      &skuListResp,
		ImageList:    &imageListResp,
		CreatedAt:    spu.CreatedAt,
		UpdatedAt:    spu.UpdatedAt,
	}
	return &spuResp, nil
}

// UserGetProduct 用户端商品详情
// 公开接口，返回已上架商品的详细信息（含SKU列表和图片，不含成本价/锁库存等内部字段）
// 接收值：spuId - 商品SPU ID
// 返回值：*response.UserGetProductResp - 用户端商品详情响应（仅返回已上架且未删除的商品）
//
//	error - 错误信息
func (p *ProductService) UserGetProduct(spuId int64) (*response.UserGetProductResp, error) {
	var spuResp response.UserGetProductResp
	var skuListResp []response.UserSkuList
	var imageListResp []response.ImageList

	// 调用数据层根据spuID获取spu信息
	spu, err := p.ProductRepo.GetSpuById(spuId)
	if err != nil {
		return nil, err
	}
	// 过滤商品信息，已删除和未上架的不予返回
	if spu == nil || spu.IsDeleted || spu.SpuStatus != "published" {
		return nil, nil
	}

	// 调用数据层获取spu对应的类目信息
	category, err := p.CategoryRepo.GetCategoryById(spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, model.CategoryNotExist
	}
	// 调用数据层获取spu下面全部的sku列表信息
	skuList, err := p.ProductRepo.GetSkuListBySpuId(spuId)
	if err != nil {
		return nil, err
	}
	// 调用数据层获取spu下面全部的图像列表信息
	imageList, err := p.ProductRepo.GetImageListBySpuId(spuId)
	if err != nil {
		return nil, err
	}

	// 根据sku列表信息 逐条构建用户端响应信息
	for _, sku := range skuList {
		skuResp := response.UserSkuList{
			SkuId:     sku.SkuId,
			SpuId:     sku.SpuId,
			SkuName:   sku.SkuName,
			SpecValue: sku.SpecValues,
			Price:     sku.Price,
			Stock:     sku.Stock,
			SoldCount: sku.SoldCount,
			SkuCode:   sku.SkuCode,
			SkuImage:  sku.SkuImage,
			SkuStatus: sku.SkuStatus,
		}
		skuListResp = append(skuListResp, skuResp)
	}

	// 根据图像列表信息 逐条构建用户端响应信息
	for _, image := range imageList {
		imageResp := response.ImageList{
			ImageId:   image.ImageId,
			ImageUrl:  image.ImageUrl,
			SortOrder: image.SortOrder,
			IsMain:    image.IsMain,
		}
		imageListResp = append(imageListResp, imageResp)
	}

	// 整合全部以获取信息，构建完整用户端响应体信息
	spuResp = response.UserGetProductResp{
		SpuId:        spu.SpuId,
		SpuName:      spu.SpuName,
		CategoryId:   category.CategoryId,
		CategoryName: category.CategoryName,
		Brand:        spu.Brand,
		Description:  spu.Description,
		MainImage:    spu.MainImage,
		SpecTemplate: spu.SpecTemplate,
		SpuStatus:    spu.SpuStatus,
		Priority:     spu.Priority,
		SkuList:      &skuListResp,
		ImageList:    &imageListResp,
		CreatedAt:    spu.CreatedAt,
		UpdatedAt:    spu.UpdatedAt,
	}
	return &spuResp, nil
}

// UpdateProduct 部分更新商品
// 部分更新商品SPU基本字段，不修改SKU和图片。仅允许withdrawn→draft状态转换
// 接收值：spuId - 商品SPU ID
//
//	updateSpu - 待更新的字段映射（支持spu_name/category_id/brand/description/main_image/priority/spu_status）
//
// 返回值：*response.GetProductResp - 更新后的商品完整信息
//
//	error - 错误信息
func (p *ProductService) UpdateProduct(spuId int64, updateSpu map[string]interface{}) (*response.GetProductResp, error) {
	err := p.db.Transaction(func(tx *gorm.DB) error {

		// 构建事务实例
		spuTxRepo := p.ProductRepo.WithTx(tx)

		// 获取待更新spu信息
		oldSpu, err := spuTxRepo.GetSpuById(spuId)
		if err != nil {
			return err
		}
		if oldSpu == nil {
			return model.ProductNotExist
		}

		// 规格变更校验：已上架的商品不能修改规格信息
		if updateSpu["spec_template"] != nil && oldSpu.SpuStatus == "published" {
			return model.ErrPublishedCantChangeSpec
		}
		// 状态转换校验：更新接口只允许 withdrawn→draft
		if newStatus, ok := updateSpu["spu_status"].(string); ok && newStatus != oldSpu.SpuStatus {
			if !(oldSpu.SpuStatus == "withdrawn" && newStatus == "draft") {
				return model.ErrInvalidStatusTransition
			}
		}

		// 类目变更校验：目标类目必须存在且为叶子节点
		if catId, ok := updateSpu["category_id"].(float64); ok {
			category, err := p.CategoryRepo.GetCategoryById(int64(catId))
			if err != nil {
				return err
			}
			if category == nil || !category.IsLeaf || category.Status != "active" {
				return model.ErrCategoryNotUsed
			}
		}

		// 用传入的更新信息 更新商品信息
		newSpu := *oldSpu

		config := &mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &newSpu,
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return err
		}
		if err := decoder.Decode(updateSpu); err != nil {
			return err
		}

		// 调用数据层更新商品信息
		if err := spuTxRepo.UpdateSpu(&newSpu); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 返回更新后的商品信息
	return p.GetProduct(spuId)
}

// UpdateProductFull 全量更新商品
// 全量更新SPU信息并同步修改SKU和图片列表。传入SKU/图片列表与数据库对比，有sku_id则更新，无则新增，数据库中已有但不在传入列表中的则软删除
// 接收值：spuId - 商品SPU ID
//
//	req - 全量更新请求（含SPU基本信息、SKU列表、图片列表、待删除图片ID列表）
//
// 返回值：*response.GetProductResp - 更新后的商品完整信息
//
//	error - 错误信息
func (p *ProductService) UpdateProductFull(spuId int64, req requset.FullUpdateProductReq) (*response.GetProductResp, error) {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		// 创建事务实例
		spuTxRepo := p.ProductRepo.WithTx(tx)

		// 校验商品存在
		oldSpu, err := spuTxRepo.GetSpuById(spuId)
		if err != nil {
			return err
		}
		if oldSpu == nil {
			return gorm.ErrRecordNotFound
		}

		// 已上架商品不允许修改规格模板
		if oldSpu.SpuStatus == "published" && req.SpecTemplate != nil {
			return model.ErrPublishedCantChangeSpec
		}

		// 更新 SPU 基本信息
		newSpu := *oldSpu
		if req.SpuName != "" {
			newSpu.SpuName = req.SpuName
		}
		if req.CategoryId != nil {
			newSpu.CategoryId = *req.CategoryId
		}
		if req.Brand != "" {
			newSpu.Brand = req.Brand
		}
		if req.Description != "" {
			newSpu.Description = req.Description
		}
		if req.MainImage != "" {
			newSpu.MainImage = req.MainImage
		}
		if req.SpecTemplate != nil {
			newSpu.SpecTemplate = *req.SpecTemplate
		}
		if req.Priority != nil {
			newSpu.Priority = *req.Priority
		}
		if err := spuTxRepo.UpdateSpu(&newSpu); err != nil {
			return err
		}

		// 处理 SKU 列表：对比增/改/删
		if req.SkuList != nil {
			if len(*req.SkuList) == 0 {
				return model.ErrSkuListEmpty
			}

			// 规格与 SKU 校验（规则 1-4）
			if err := validateSpecSku(spuId, newSpu.SpecTemplate, *req.SkuList, spuTxRepo, false); err != nil {
				return err
			}

			// 收集请求中的已有 SKU ID
			keepIds := make([]int64, 0, len(*req.SkuList))
			for _, sku := range *req.SkuList {
				if sku.SkuId > 0 {
					keepIds = append(keepIds, sku.SkuId)
				}
			}

			// 软删不在请求中的 SKU
			if err := spuTxRepo.SoftDeleteSkusExcept(spuId, keepIds); err != nil {
				return err
			}

			// 增/改 SKU
			for _, sku := range *req.SkuList {
				sku := sku
				sku.SpuId = spuId
				if sku.SkuId > 0 {
					if err := spuTxRepo.UpdateSku(&sku); err != nil {
						return err
					}
				} else {
					if err := spuTxRepo.CreateSku(&sku); err != nil {
						return err
					}
				}
			}
		}

		// 处理图片：删/增/改
		if err := spuTxRepo.DeleteImageByIds(req.DeleteImageIds); err != nil {
			return err
		}
		if req.ImageList != nil {
			for _, img := range *req.ImageList {
				img := img
				img.SpuId = spuId
				if img.ImageId > 0 {
					if err := spuTxRepo.UpdateImage(&img); err != nil {
						return err
					}
				} else {
					if err := spuTxRepo.CreateImage(&img); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	// 返回更新后的商品信息
	return p.GetProduct(spuId)
}

// DeleteProduct 删除商品
// 软删除商品，同时级联软删除所有SKU。
// 接收值：spuId - 商品SPU ID
// 返回值：error - 错误信息
// TODO：订单关联检查
func (p *ProductService) DeleteProduct(spuId int64) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		// 创建事务实例
		spuTxRepo := p.ProductRepo.WithTx(tx)
		// 校验待删除商品是否存在
		spu, err := spuTxRepo.GetSpuById(spuId)
		if err != nil {
			return err
		}
		if spu == nil {
			return model.ProductNotExist
		}

		// 批量软删除此spu下的sku
		if err := spuTxRepo.BatchDeleteSkuById(spuId); err != nil {
			return err
		}
		// 调用数据层软删除spu
		if err := spuTxRepo.DeleteSpuById(spuId); err != nil {
			return err
		}
		return nil
	})
}

// PublishProduct 上架商品
// 将草稿或已下架商品上架，上架前校验类目有效性和至少有一个可售SKU（启用且库存>0）
// 接收值：spuId - 商品SPU ID
// 返回值：error - 错误信息
func (p *ProductService) PublishProduct(spuId int64) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		// 创新事务实例
		spuTxRepo := p.ProductRepo.WithTx(tx)

		// 校验待上架商品是否存在
		spu, err := spuTxRepo.GetSpuById(spuId)
		if err != nil {
			return err
		}
		if spu == nil {
			return model.ProductNotExist
		}

		// 获取待上架商品所属类目的信息
		category, err := p.CategoryRepo.GetCategoryById(spu.CategoryId)
		if err != nil {
			return err
		}
		if category == nil {
			return model.ErrCategoryNotUsed
		}

		// 校验待上架商品所属类目是否为叶子节点且处于可用状态
		if !(category.IsLeaf && category.Status == "active") {
			return model.ErrCategoryNotUsed
		}

		// 校验待上架商品是否处于可上架状态，存在库存>0，价格>0的sku信息
		if err := p.ProductRepo.ValidateSpuPublish(spuId); err != nil {
			return err
		}

		// 调用数据层完成上架操作
		if err := spuTxRepo.PublishProduct(spuId); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ErrInvalidStatusTransition
			}
		}
		return nil
	})
}

// WithdrawProduct 下架商品
// 将已上架商品下架，用户端不再展示。仅published状态可下架
// 接收值：spuId - 商品SPU ID
// 返回值：error - 错误信息
func (p *ProductService) WithdrawProduct(spuId int64) error {

	// 校验待上架商品是否存在
	spu, err := p.ProductRepo.GetSpuById(spuId)
	if err != nil {
		return err
	}
	if spu == nil {
		return model.ProductNotExist
	}

	// 调用数据层完成下架操作
	if err := p.ProductRepo.WithdrawProduct(spuId); err != nil {
		return err
	}
	return nil
}
