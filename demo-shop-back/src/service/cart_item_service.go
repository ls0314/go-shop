package service

import (
	"demo-shop-back/db"
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

// CartItemService 购物车服务层实例
type CartItemService struct {
	CartItemRepo *repository.CartItemRepo
	ProductRepo  *repository.ProductRepo
	db           *gorm.DB
}

// NewCartItemService 创建购物车服务层实例
// 接收值：注入的数据库连接（由 composition root 提供）
// 返回值：*CartItemService - 购物车服务层实例指针
func NewCartItemService() *CartItemService {
	return &CartItemService{
		CartItemRepo: repository.NewCartItemRepo(db.DB),
		ProductRepo:  repository.NewProductRepo(db.DB),
		db:           db.DB,
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
	// 根据skuId获得sku对应数据
	sku, err := ci.ProductRepo.GetSku(skuId)
	if err != nil {
		return nil, err
	}
	// 确保sku处于可用的状态下
	if sku.SkuStatus != model.SkuStatusActive || sku.IsDeleted {
		return nil, model.ErrSkuDisabled
	}
	// 根据sku对应的spuId获取spu信息
	spu, err := ci.ProductRepo.GetSpuById(sku.SpuId)
	if err != nil {
		return nil, err
	}
	// 确保spu已上架
	if spu.SpuStatus != model.SpuStatusPublished || spu.IsDeleted {
		return nil, model.ErrSpuDisabled
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
	// 调用数据层获取当前用户的购物车列表
	cartItemList, err := ci.CartItemRepo.GetCartItemList(userId)
	if err != nil {
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

	// 返回更新后的购物车信息
	return ci.CartItemRepo.GetCartItemResp(userId, cartItem.SkuId)
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
