package productservicelogic

import (
	"context"
	"demo-shop/services/product/internal/infra/es"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/repository"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
)

// ============================================================
// 商品域公共件:错误归类、缓存、实时库存、ES 同步、规格校验
// 从单体 src/service/product_service.go 平移,沿途修复以 [FIX-n] 标注。
// ============================================================

// 商品域可预期的业务失败。落在此集合内以 error_msg 返回(gRPC error 为 nil),
// 调用方(单体/BFF)按文案还原成本地错误变量;其余视为基础设施故障,
// gRPC error 非 nil 由调用方重试。与类目域 categoryhelper.go 同一约定。
var productBizErrors = []error{
	model.ProductNotExist,
	model.ErrSpuDisabled,
	model.ErrSpuTemplate,
	model.ErrSkuNum,
	model.ErrSpecValues,
	model.ErrSkuCodeNotOnly,
	model.ErrSpecValuesNotOnly,
	model.ErrNoActiveSku,
	model.ErrNoAvailableStock,
	model.ErrInvalidPrice,
	model.ErrCategoryNotUsed,
	model.ErrPublishedCantChangeSpec,
	model.ErrSkuListEmpty,
	model.ErrInvalidStatusTransition,
}

func isProductBizError(err error) bool {
	for _, target := range productBizErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// specItem 规格模板中的单条规格
type specItem struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// contains 判断切片是否包含目标字符串
func contains(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// validateSpecSku 校验规格模板与 SKU 列表。
// 前半段(规则 1-3)是纯内存校验,拆到 validateSpecSkuRules 以便单测;
// 后半段(规则 4)需要查库,sqlmock 成本高于收益,故只对纯逻辑做用例覆盖。
func validateSpecSku(spuId int64, specTemplate interface{}, skuList []model.SysProductSku, repo *repository.ProductRepo, checkSpecValue bool) error {
	if _, err := validateSpecSkuRules(specTemplate, skuList); err != nil {
		return err
	}

	// 规则4:规格组合唯一(仅创建路径需要查库)
	if !checkSpecValue {
		return nil
	}
	seen := make(map[string]bool, len(skuList))
	for _, sku := range skuList {
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
	return nil
}

// validateSpecSkuRules 纯内存校验,返回解析出的规格模板。
//
// 规则1 模板格式合法(每条规格有名字且有取值);
// 规则2 SKU 数 = 各规格取值数的笛卡尔积;
// 规则3 每个 SKU 的 spec_values 键值必须与模板逐项匹配。
func validateSpecSkuRules(specTemplate interface{}, skuList []model.SysProductSku) ([]specItem, error) {
	b, err := json.Marshal(specTemplate)
	if err != nil {
		return nil, model.ErrSpuTemplate
	}
	var specs []specItem
	if err := json.Unmarshal(b, &specs); err != nil || len(specs) == 0 {
		return nil, model.ErrSpuTemplate
	}

	// 规则1:规格模板格式校验
	for _, s := range specs {
		if s.Name == "" || len(s.Values) == 0 {
			return nil, model.ErrSpuTemplate
		}
	}

	// 规则2:SKU 数量 = 笛卡尔积
	expected := 1
	for _, s := range specs {
		expected *= len(s.Values)
	}
	if len(skuList) != expected {
		return nil, model.ErrSkuNum
	}

	// 规则3:逐 SKU 校验规格键值
	for _, sku := range skuList {
		if len(sku.SpecValues) != len(specs) {
			return nil, model.ErrSpecValues
		}
		for _, s := range specs {
			val, ok := sku.SpecValues[s.Name]
			if !ok {
				return nil, model.ErrSpecValues
			}
			if !contains(s.Values, fmt.Sprint(val)) {
				return nil, model.ErrSpecValues
			}
		}
	}

	return specs, nil
}

// normalizePage 统一分页参数:page<=0 → 1;pageSize<=0 → 10;超过 maxSize 则封顶。
// 管理端与用户端用不同上限,故由上调用方传入。
func normalizePage(page, pageSize, maxSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return page, pageSize
}

// nullableInt64 把 proto 的 optional int64(wrappers.Int64Value)转成查询条件。
// 未传(如 category_id 留空表示不过滤)返回 nil,已传则取指针指向的值 ——
// 不能直接用 in.GetCategoryId().GetValue(),那样"传了 0"与"没传"无法区分,
// 也不能区分 nil 解引用。
func nullableInt64(v *wrapperspb.Int64Value) *int64 {
	if v == nil {
		return nil
	}
	val := v.GetValue()
	return &val
}

// ============================================================
// 缓存
// ============================================================

// getCategory 取类目信息:缓存优先,miss 查 DB 回写。仅用于展示读,写路径校验请直查 DB
func getCategory(svcCtx *svc.ServiceContext, categoryId int64) (*model.SysCategory, error) {
	key := fmt.Sprintf("category:%d", categoryId)
	if svcCtx.Redis != nil {
		var cat model.SysCategory
		hit, err := utils.GetJSONCtx(context.Background(), svcCtx.Redis, key, &cat)
		if err == nil && hit {
			return &cat, nil
		}
	}
	cat, err := svcCtx.CategoryRepo.GetCategoryById(categoryId)
	if err != nil || cat == nil {
		return cat, err
	}
	if svcCtx.Redis != nil { // 弱依赖降级:Redis 未配置时跳过回写,不影响返回
		_ = utils.SetJSONCtx(context.Background(), svcCtx.Redis, key, cat, 3600)
	}
	return cat, nil
}

// setProductDetailCache 写商品详情缓存
func setProductDetailCache(svcCtx *svc.ServiceContext, scene string, spuId int64, v interface{}) {
	if svcCtx.Redis == nil {
		return
	}
	_ = utils.SetJSONCtx(context.Background(), svcCtx.Redis,
		fmt.Sprintf("product:detail:%s:%d", scene, spuId), v, 600)
}

// getProductDetailCache 读商品详情缓存
func getProductDetailCache(svcCtx *svc.ServiceContext, scene string, spuId int64, dest interface{}) bool {
	if svcCtx.Redis == nil {
		return false
	}
	hit, err := utils.GetJSONCtx(context.Background(), svcCtx.Redis,
		fmt.Sprintf("product:detail:%s:%d", scene, spuId), dest)
	return err == nil && hit
}

// delProductDetailCache 失效商品详情的两个场景缓存(管理端/用户端)
func delProductDetailCache(svcCtx *svc.ServiceContext, spuId int64) {
	if svcCtx.Redis == nil {
		return
	}
	_, _ = svcCtx.Redis.Del(
		fmt.Sprintf("product:detail:admin:%d", spuId),
		fmt.Sprintf("product:detail:user:%d", spuId),
	)
}

// ============================================================
// 实时库存(闸门键)
// ============================================================

// liveStock 取单个 SKU 的实时可售库存。
//
// 口径取自库存闸门键 sku:stock:{id} —— 该键由库存四操作(下单扣/释放/退款)
// 维护并在变更后失效,是"当前可售量"的权威值;未命中则回落到 DB 并回填该键
// (与 stock_reconcile 对账任务的收敛方向一致)。
//
// 注意:管理端与用户端详情缓存存的都是 DB 快照,两侧都必须调用本函数覆盖,
// 否则同一商品在两个端上会显示不同库存(停留在写缓存那一刻的值)。
func liveStock(svcCtx *svc.ServiceContext, skuId int64, dbStock int64) (int64, bool) {
	if svcCtx.Redis == nil || skuId == 0 {
		return dbStock, false
	}
	stockKey := utils.StockGateKey(skuId)
	if val, err := svcCtx.Redis.Get(stockKey); err == nil {
		if stock, perr := strconv.ParseInt(val, 10, 64); perr == nil {
			return stock, true
		}
	}

	// 缓存未命中:以 DB 为准回填闸门键
	sku, err := svcCtx.ProductRepo.GetSku(skuId)
	if err != nil || sku == nil {
		return dbStock, false
	}
	_ = svcCtx.Redis.Setex(stockKey, strconv.FormatInt(sku.Stock, 10), 30)
	return dbStock, false
}

// fillUserSkuStock 用实时库存覆盖用户端 SKU 列表
func fillUserSkuStock(svcCtx *svc.ServiceContext, skuList *[]model.UserSkuList) {
	if skuList == nil {
		return
	}
	for i := range *skuList {
		if stock, ok := liveStock(svcCtx, (*skuList)[i].SkuId, (*skuList)[i].Stock); ok {
			(*skuList)[i].Stock = stock
		}
	}
}

// fillAdminSkuStock 用实时库存覆盖管理端 SKU 列表。
// 管理端额外展示 lock_stock,但可售库存与用户端必须同源,故一并覆盖 stock。
func fillAdminSkuStock(svcCtx *svc.ServiceContext, skuList *[]model.SkuList) {
	if skuList == nil {
		return
	}
	for i := range *skuList {
		if stock, ok := liveStock(svcCtx, (*skuList)[i].SkuId, (*skuList)[i].Stock); ok {
			(*skuList)[i].Stock = stock
		}
	}
}

// ============================================================
// ES 索引同步(尽力而为)
// ============================================================

// syncProductToES 把单个 SPU 同步到搜索索引。失败只告警,不影响主流程:
// 索引落后最多让用户端搜索少一条,而回滚商品写入的代价更大。
func syncProductToES(logger logx.Logger, svcCtx *svc.ServiceContext, spuId int64) {
	if svcCtx.ES == nil {
		return // ES 未启用,静默跳过
	}
	doc, err := svcCtx.ProductRepo.GetSpuESDoc(spuId)
	if err != nil {
		logger.Errorf("查询 ES 文档数据失败 spuId=%d: %v", spuId, err)
		return
	}
	// [FIX-11] 单体此处依赖 GetSpuESDoc 返回 ErrRecordNotFound 来跳过零值文档,
	// 但 Scan 对空结果只留零值、不返回该错误,导致 SPU 不存在时会索引一条 spu_id=0 的垃圾。
	// 改为显式判空。
	if doc == nil || doc.SpuId == 0 {
		logger.Errorf("跳过 ES 索引:未查到 SPU spuId=%d", spuId)
		return
	}
	if err := svcCtx.ES.IndexProduct(context.Background(), toESProduct(doc)); err != nil {
		logger.Errorf("ES 索引失败 spuId=%d: %v", spuId, err)
	}
}

// removeProductFromES 从搜索索引删除(尽力而为,失败只告警)
func removeProductFromES(logger logx.Logger, svcCtx *svc.ServiceContext, spuId int64) {
	if svcCtx.ES == nil {
		return
	}
	if err := svcCtx.ES.DeleteProduct(context.Background(), spuId); err != nil {
		logger.Errorf("ES 删除失败 spuId=%d: %v", spuId, err)
	}
}

// toESProduct 领域文档 → 搜索文档
func toESProduct(doc *model.SpuESDoc) *es.ESProduct {
	return &es.ESProduct{
		SpuId:        doc.SpuId,
		SpuName:      doc.SpuName,
		Brand:        doc.Brand,
		Description:  doc.Description,
		CategoryId:   doc.CategoryId,
		CategoryName: doc.CategoryName,
		MainImage:    doc.MainImage,
		TotalStock:   doc.TotalStock,
		TotalSold:    doc.TotalSold,
		Priority:     doc.Priority,
		SpuStatus:    doc.SpuStatus,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		MinPrice:     doc.MinPrice,
		MaxPrice:     doc.MaxPrice,
	}
}

// ============================================================
// 详情组装
// ============================================================

// assembleAdminDetail 组装管理端详情。
// 落库的缓存里存的是 DB 快照,可售库存以闸门键为准,组装时覆盖一次。
func assembleAdminDetail(svcCtx *svc.ServiceContext, spu *model.SysProductSpu) (*model.GetProductResp, error) {
	category, err := getCategory(svcCtx, spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, model.CategoryNotExist
	}

	skuList, err := svcCtx.ProductRepo.GetSkuListBySpuId(spu.SpuId)
	if err != nil {
		return nil, err
	}
	imageList, err := svcCtx.ProductRepo.GetImageListBySpuId(spu.SpuId)
	if err != nil {
		return nil, err
	}

	skuListResp := make([]model.SkuList, 0, len(skuList))
	for _, sku := range skuList {
		skuListResp = append(skuListResp, model.SkuList{
			SkuId:      sku.SkuId,
			SpuId:      sku.SpuId,
			SkuName:    sku.SkuName,
			SpecValues: sku.SpecValues,
			Price:      sku.Price,
			CostPrice:  sku.CostPrice,
			Stock:      sku.Stock,
			LockStock:  sku.LockStock,
			SoldCount:  sku.SoldCount,
			SkuCode:    sku.SkuCode,
			SkuImage:   sku.SkuImage,
			SkuStatus:  sku.SkuStatus,
		})
	}

	imageListResp := make([]model.ImageList, 0, len(imageList))
	for _, image := range imageList {
		imageListResp = append(imageListResp, model.ImageList{
			ImageId:   image.ImageId,
			ImageUrl:  image.ImageUrl,
			SortOrder: image.SortOrder,
			IsMain:    image.IsMain,
		})
	}

	fillAdminSkuStock(svcCtx, &skuListResp)

	return &model.GetProductResp{
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
	}, nil
}

// assembleUserDetail 组装用户端详情,口径同管理端(库存同源),字段按端裁剪
func assembleUserDetail(svcCtx *svc.ServiceContext, spu *model.SysProductSpu) (*model.UserGetProductResp, error) {
	category, err := getCategory(svcCtx, spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, model.CategoryNotExist
	}
	skuList, err := svcCtx.ProductRepo.GetSkuListBySpuId(spu.SpuId)
	if err != nil {
		return nil, err
	}
	imageList, err := svcCtx.ProductRepo.GetImageListBySpuId(spu.SpuId)
	if err != nil {
		return nil, err
	}

	skuListResp := make([]model.UserSkuList, 0, len(skuList))
	for _, sku := range skuList {
		skuListResp = append(skuListResp, model.UserSkuList{
			SkuId:      sku.SkuId,
			SpuId:      sku.SpuId,
			SkuName:    sku.SkuName,
			SpecValues: sku.SpecValues,
			Price:      sku.Price,
			Stock:      sku.Stock,
			SoldCount:  sku.SoldCount,
			SkuCode:    sku.SkuCode,
			SkuImage:   sku.SkuImage,
			SkuStatus:  sku.SkuStatus,
		})
	}

	imageListResp := make([]model.ImageList, 0, len(imageList))
	for _, image := range imageList {
		imageListResp = append(imageListResp, model.ImageList{
			ImageId:   image.ImageId,
			ImageUrl:  image.ImageUrl,
			SortOrder: image.SortOrder,
			IsMain:    image.IsMain,
		})
	}

	fillUserSkuStock(svcCtx, &skuListResp)

	return &model.UserGetProductResp{
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
	}, nil
}

// loadAdminDetail 管理端详情:缓存优先,miss 查库组装并回写
func loadAdminDetail(svcCtx *svc.ServiceContext, spuId int64) (*model.GetProductResp, error) {
	var cached model.GetProductResp
	if getProductDetailCache(svcCtx, "admin", spuId, &cached) {
		fillAdminSkuStock(svcCtx, cached.SkuList)
		return &cached, nil
	}

	spu, err := svcCtx.ProductRepo.GetSpuById(spuId)
	if err != nil {
		// [FIX-13] 单体在这里返回 (nil, nil),把"查不到"当成功,调用方拿到 nil 响应。
		// 统一改为业务错误,与 UpdateProduct/DeleteProduct/PublishProduct 一致。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ProductNotExist
		}
		return nil, err
	}

	resp, err := assembleAdminDetail(svcCtx, spu)
	if err != nil {
		return nil, err
	}
	setProductDetailCache(svcCtx, "admin", spuId, resp)
	return resp, nil
}

// loadUserDetail 用户端详情:缓存优先,miss 查库组装并回写
func loadUserDetail(svcCtx *svc.ServiceContext, spuId int64) (*model.UserGetProductResp, error) {
	var cached model.UserGetProductResp
	if getProductDetailCache(svcCtx, "user", spuId, &cached) {
		fillUserSkuStock(svcCtx, cached.SkuList)
		return &cached, nil
	}

	spu, err := svcCtx.ProductRepo.GetSpuById(spuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ProductNotExist
		}
		return nil, err
	}
	if spu.IsDeleted || spu.SpuStatus != model.SpuStatusPublished {
		return nil, model.ProductNotExist
	}

	resp, err := assembleUserDetail(svcCtx, spu)
	if err != nil {
		return nil, err
	}
	setProductDetailCache(svcCtx, "user", spuId, resp)
	return resp, nil
}
