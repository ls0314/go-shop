package service

import (
	"demo-shop-back/src/infra/productclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ============================================================
//	定义及实例化
// ============================================================

// CartItemService 购物车服务层实例。
//
// ProductRepo 已摘除:商品/SKU 表的所有权在 product-service 的库里,
// 加购校验改经 ProductRPC 做 —— 否则本地读的是另一个库的商品状态。
type CartItemService struct {
	CartItemRepo *repository.CartItemRepo
	ProductRPC   *productclient.ProductClient
	db           *gorm.DB
}

// NewCartItemService 创建购物车服务层实例
// 接收值：注入的数据库连接（由 composition root 提供）
// 返回值：*CartItemService - 购物车服务层实例指针
func NewCartItemService(deps ServiceDeps) *CartItemService {
	return &CartItemService{
		CartItemRepo: repository.NewCartItemRepo(deps.DB),
		ProductRPC:   deps.ProductRPC,
		db:           deps.DB,
	}
}

// ============================================================
// 规则校验
// ============================================================

// validateQuantity 校验购买数 （合法的购买数量的区间为1~999）
// 接收值：quantity - 购买数
// 返回值：error - 错误信息
func (ci *CartItemService) validateQuantity(quantity int64) error {
	// 确保购买数在1~999这一区间内
	if quantity < model.CartItemMinQuantity || quantity > model.CartItemMaxQuantity {
		return model.QuantityMax
	}
	return nil
}

// ValidateProductAvailable 校验商品状态（保证商品可以被正常购买）
// 接收值：skuId - 商品skuId
// 返回值：
//
//	*model.SysProductSku - 对应的商品Sku结构体指针
//	error - 错误信息
func (ci *CartItemService) validateProductAvailable(skuId int64) (*model.SysProductSku, error) {
	// 可购买判定在服务端做(SKU active + 未删,SPU published + 未删三重条件)
	ok, errMsg, err := ci.ProductRPC.IsSkuPurchasable(skuId)
	if err != nil {
		return nil, err
	}
	if !ok {
		// errMsg 是服务端给的具体原因(SKU 禁用 / 商品下架),直接透出给前端
		if errMsg == model.ErrSpuDisabled.Error() {
			return nil, model.ErrSpuDisabled
		}
		return nil, model.ErrSkuDisabled
	}

	// 校验通过后再取实体:购物车要落 sku 的价格与名称快照
	sku, errMsg, err := ci.ProductRPC.GetSku(skuId)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, productclient.RestoreError(errMsg)
	}
	return sku, nil
}

// ============================================================
// 购物车模块业务操作
// ============================================================

// createNewCartItem 创建新购物车（内部方法）
// 接收值:
//
//	cartItem - 待创建的购物车信息
//	stock - sku的库存数量
//
// 返回值:
//
//	*response.CartItemCreateResp - 创建接口响应体（包含购物车ID和购买数）
//	error - 错误信息
func (ci *CartItemService) createNewCartItem(cartItem *model.UserCartItem, stock int64) (*response.CartItemCreateResp, error) {
	// 调用数据层获取当前用户购物车总数
	total, err := ci.CartItemRepo.GetCartItemTotal(cartItem.UserId)
	if err != nil {
		return nil, err
	}
	// 确保每个用户的购物车数量不超过100
	if total >= model.CartMaxItemCount {
		return nil, model.CartItemMax
	}
	// 确保库存充足
	if cartItem.Quantity > stock {
		return nil, model.ErrStockNotEnough
	}
	// 调用数据层创建购物车
	if err := ci.CartItemRepo.CreateCartItem(cartItem); err != nil {
		return nil, err
	}

	// 返回最终响应
	return &response.CartItemCreateResp{
		CartItemId: cartItem.CartItemId,
		Quantity:   cartItem.Quantity,
	}, nil
}

// updateExistCartItem 更新已有购物车（内部方法）
// 接收值:
//
//	existItem - 已有的购物车指针
//	addQuantity - 新添加的购买数量
//	stock - sku的库存数量
//
// 返回值:
//
//	*response.CartItemCreateResp - 创建接口响应体（包含购物车ID和购买数）
//	error - 错误信息
func (ci *CartItemService) updateExistCartItem(existItem *response.CartItemListResp, addQuantity, stock int64) (*response.CartItemCreateResp, error) {
	// 总购买数 = 已存在购物车的购买数 + 新增购买数
	totalQuantity := existItem.Quantity + addQuantity

	// 校验相关规则
	switch {
	// 确保购买数小于999
	case totalQuantity > model.CartItemMaxQuantity:
		return nil, model.QuantityMax
	//	确保购买数小于库存
	case totalQuantity > stock:
		return nil, model.ErrStockNotEnough
	}
	// 调用数据层更新购物车购买数
	if err := ci.CartItemRepo.UpdateCartItemQuantity(existItem.CartItemId, totalQuantity); err != nil {
		return nil, err
	}
	// 返回响应数据
	return &response.CartItemCreateResp{
		CartItemId: existItem.CartItemId,
		Quantity:   totalQuantity,
	}, nil
}

// CreateCartItem 加入购物车
// 接收值:
//
//	userId - 用户ID
//	cartItem - 待添加购物车指针
//
// 返回值:
//
//	*response.CartItemCreateResp - 创建接口响应体（包含购物车ID和购买数）
//	error - 错误信息
func (ci *CartItemService) CreateCartItem(userId int64, cartItem *model.UserCartItem) (*response.CartItemCreateResp, error) {
	// 若购物车结构体指针内未传入用户Id，则将当前用户Id传入购物车结构体
	if userId != cartItem.UserId {
		return nil, model.NotAuthority
	}
	// 校验购买数量是否合规
	if err := ci.validateQuantity(cartItem.Quantity); err != nil {
		return nil, err
	}
	// 校验商品是否可购买
	sku, err := ci.validateProductAvailable(cartItem.SkuId)
	if err != nil {
		return nil, err
	}
	// 调用数据层获取当前用户对此商品有购物车记录
	existItem, err := ci.CartItemRepo.GetCartItemResp(cartItem.UserId, cartItem.SkuId)
	if err != nil && !errors.Is(err, model.CartItemNotExist) {
		return nil, err
	}
	// 若不存在则新建购物车，已有就更新购物车信息
	if existItem == nil {
		return ci.createNewCartItem(cartItem, sku.Stock)
	} else {
		return ci.updateExistCartItem(existItem, cartItem.Quantity, sku.Stock)
	}
}

// fillCartItemProducts 用 product-service 的商品数据回填购物车条目的商品字段。
//
// 分库后 sys_product_sku / sys_product_spu 归 product-service,
// 原来仓库层那条"cart_item LEFT JOIN sku LEFT JOIN spu"跨库 JOIN 已不可用,
// 改为"本地查购物车 + RPC 批量取商品"两步,在服务层拼装。
//
// 拼装口径与拆分前逐字段对齐(前端零感知):
//   - SKU 查不到 → 商品字段留空,由 calculateCartItem 判为不可购买
//     (提交前是 LEFT JOIN,查不到同样留空);
//   - 库存取 DB 快照(与 LEFT JOIN sku.stock 一致),不是闸门实时值 ——
//     购物车页展示的是"账面库存",真正的并发扣减在下单锁定那一步。
func (ci *CartItemService) fillCartItemProducts(items []response.CartItemListResp) error {
	if len(items) == 0 {
		return nil
	}

	skuIds := make([]int64, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for i := range items {
		if items[i].SkuId == 0 {
			continue
		}
		if _, ok := seen[items[i].SkuId]; ok {
			continue
		}
		seen[items[i].SkuId] = struct{}{}
		skuIds = append(skuIds, items[i].SkuId)
	}
	if len(skuIds) == 0 {
		return nil
	}

	skuSnapshots, errMsg, err := ci.ProductRPC.BatchGetSkus(skuIds)
	if err != nil {
		return err
	}
	if errMsg != "" {
		return productclient.RestoreError(errMsg)
	}

	skuMap := make(map[int64]productclient.CartSkuSnapshot, len(skuSnapshots))
	for _, snap := range skuSnapshots {
		if snap.Sku != nil {
			skuMap[snap.Sku.SkuId] = snap
		}
	}

	for i := range items {
		snap, ok := skuMap[items[i].SkuId]
		if !ok {
			continue // SKU 已删/不存在:留空,交给 calculateCartItem 标记不可购买
		}
		items[i].SkuId = snap.Sku.SkuId
		items[i].SkuName = snap.Sku.SkuName
		items[i].SpecValues = snap.Sku.SpecValues
		items[i].SkuImage = snap.Sku.SkuImage
		items[i].SkuStatus = snap.Sku.SkuStatus
		items[i].Price = snap.Sku.Price
		items[i].Stock = snap.Sku.Stock

		items[i].SpuId = snap.SpuId
		items[i].SpuName = snap.SpuName
		items[i].MainImage = snap.SpuMainImage
		items[i].SpuStatus = snap.SpuStatus
	}

	return nil
}

// calculateCartItem 计算更新购物车列表不可购买的状态和原因（内部方法）
func (ci *CartItemService) calculateCartItem(item *response.CartItemListResp) error {
	item.Subtotal = item.Price * float64(item.Quantity)

	// 按业务优先级判定状态：SPU上架状态 > SKU启用状态 > 库存充足性
	switch {
	case item.SpuStatus != model.SpuStatusPublished:
		item.IsAvailable = false
		item.UnavailableReason = model.ProductWithdraw
	case item.SkuStatus != model.SkuStatusActive:
		item.IsAvailable = false
		item.UnavailableReason = model.ProductWithdraw
	case item.Stock < item.Quantity:
		item.IsAvailable = false
		item.UnavailableReason = fmt.Sprintf("库存不足,仅剩%d件", item.Stock)
	default:
		item.IsAvailable = true
		item.UnavailableReason = ""
	}
	return nil
}

// GetCartItemList 获取当前用户的购物车列表（包含不可购买项和不可购买原因）
// 接收值： userId - 当前用户Id
// 返回值：
//
//	[]response.CartItemListResp - 当前用户的购物车列表
//	error - 错误信息
func (ci *CartItemService) GetCartItemList(userId int64) ([]response.CartItemListResp, error) {
	// 调用数据层获取当前用户购物车行(user_cart_item 单表)
	cartItemList, err := ci.CartItemRepo.GetCartItemList(userId)
	if err != nil {
		return nil, err
	}
	// 经 product-service 回填商品字段(原跨库 JOIN 的替代)
	if err := ci.fillCartItemProducts(cartItemList); err != nil {
		return nil, err
	}
	// 逐个判断其是否可购买，若不可购买补充原因
	for i := range cartItemList {
		err := ci.calculateCartItem(&cartItemList[i])
		if err != nil {
			return nil, err
		}
	}
	return cartItemList, nil
}

// UpdateCartItem 更新购物车信息（选中状态和购买数量）
// 接收值：
//
//	cartItemId - 购物车Id
//	userId - 当前用户Id
//	updateReq - 更新参数（包含选中状态和购买数量）
//
// 返回值：
//
//	*response.CartItemListResp - 更新后的购物车信息
//	error - 错误信息
func (ci *CartItemService) UpdateCartItem(cartItemId, userId int64, updateReq requset.UpdateCartItemReq) (*response.CartItemListResp, error) {
	// 调用数据层根据购物车Id获取购物车信息判断其是否存在
	cartItem, err := ci.CartItemRepo.GetCartItem(cartItemId)
	if err != nil {
		return nil, err
	}
	// 校验归属
	if userId != cartItem.UserId {
		return nil, model.NotAuthority
	}

	// 若更新参数中为传入购买数量则将已有购买数量传入更新参数，若传入购买数量则校验购买数量
	if updateReq.Quantity == 0 {
		updateReq.Quantity = cartItem.Quantity
	} else {
		if err := ci.validateQuantity(updateReq.Quantity); err != nil {
			return nil, err
		}
	}

	// 校验库存
	if updateReq.Quantity > cartItem.Stock {
		return nil, model.ErrStockNotEnough
	}

	// 调用数据层更新购物车信息
	if err := ci.CartItemRepo.UpdateCartItem(cartItemId, updateReq); err != nil {
		return nil, err
	}

	// 返回更新后的购物车信息(经 product-service 回填商品字段)
	return ci.reloadCartItemResp(userId, cartItem.SkuId)
}

// reloadCartItemResp 重新取一条购物车并按商品字段拼装 —— 供写路径返回"更新后的完整对象"。
// 走"取 → 回填 → 单条返回"这一条路径,避免"改了切片副本却返回原值"这类易错点。
func (ci *CartItemService) reloadCartItemResp(userId, skuId int64) (*response.CartItemListResp, error) {
	updated, err := ci.CartItemRepo.GetCartItemResp(userId, skuId)
	if err != nil {
		return nil, err
	}
	list := []response.CartItemListResp{*updated}
	if err := ci.fillCartItemProducts(list); err != nil {
		return nil, err
	}
	return &list[0], nil
}

// DeleteCartItem 删除购物车
// 接收值：
//
//	cartItemId - 购物车Id
//	userId - 当前用户Id
//
// 返回值： error - 错误信息
func (ci *CartItemService) DeleteCartItem(cartItemId, userId int64) error {
	// 校验待删除购物车是否存在
	cartItem, err := ci.CartItemRepo.GetCartItem(cartItemId)
	if err != nil {
		return err
	}
	// 校验归属
	if userId != cartItem.UserId {
		return model.NotAuthority
	}
	// 调用数据层删除购物车
	return ci.CartItemRepo.DeleteCartItem(cartItemId)
}

// SelectAllCartItem 更新购物车选中状态(全选)
// 接收值：
//
//	isSelect - 选中状态
//	userId - 当前用户Id
//
// 返回值：
//
//	int64 - 被影响的购物车数量
//	error - 错误信息
func (ci *CartItemService) SelectAllCartItem(userId int64, isSelect bool) (int64, error) {
	return ci.CartItemRepo.SelectAllCartItem(userId, isSelect)
}

// GetCartItemTotal 获取当前用户的购物车总数
// 接收值： userId - 当前用户Id
//
// 返回值：
//
//	int64 - 当前用户的购物车总数
//	error - 错误信息
func (ci *CartItemService) GetCartItemTotal(userId int64) (int64, error) {
	return ci.CartItemRepo.GetCartItemTotal(userId)
}

// GetPayPreviewCartItem 获取当前用户被选中的购物车的结算预览信息
// 接收值： userId - 当前用户Id
//
// 返回值：
//
//	*response.CartItemPayResp - 结算预览购物车信息
//	error - 错误信息
func (ci *CartItemService) GetPayPreviewCartItem(userId int64) (*response.CartItemPayResp, error) {
	var items []response.CartItemListResp
	var unAvailableItem []response.CartItemListResp
	resp := response.CartItemPayResp{}

	// 调用数据层获取当前用户所有被选中的购物车列表
	cartItemList, err := ci.CartItemRepo.GetSelectCartItem(userId)
	if err != nil {
		return nil, err
	}
	// 经 product-service 回填商品字段后,下面的库存/状态判定才有依据
	if err := ci.fillCartItemProducts(cartItemList); err != nil {
		return nil, err
	}

	// 逐个校验购物车对应商品是否可购买，不可购买的写入原因
	for i := range cartItemList {
		err := ci.calculateCartItem(&cartItemList[i])
		if err != nil {
			return nil, err
		}
		// 不可购买项添加在不可购买列表，同时更新响应信息
		if cartItemList[i].IsAvailable == false {
			unAvailableItem = append(unAvailableItem, cartItemList[i])
			resp.HasUnavailable = true
		} else {
			// 可购买项添加在可购买列表，同时更新响应信息
			items = append(items, cartItemList[i])
			resp.TotalCount = resp.TotalCount + 1
			resp.TotalQuantity = resp.TotalQuantity + cartItemList[i].Quantity
			resp.TotalAmount = resp.TotalAmount + float64(cartItemList[i].Quantity)*cartItemList[i].Price
		}
	}
	// 将可购买列表和不可购买列表添加进响应信息
	resp.Items = items
	resp.UnavailableItems = unAvailableItem
	// 返回响应
	return &resp, nil
}
