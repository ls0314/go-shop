package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// 购物车实例定义
// ============================================================

// CartItemRepo 用户购物车数据层实例
type CartItemRepo struct {
	db *gorm.DB
}

// NewCartItemRepo 创建购物车表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*CartItemRepo - 购物车表数据层指针
func NewCartItemRepo() *CartItemRepo {
	return &CartItemRepo{
		db: db.DB,
	}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*CartItemRepo - 绑定事务的购物车表数据层指针
func (ci *CartItemRepo) WithTx(tx *gorm.DB) *CartItemRepo {
	return &CartItemRepo{
		db: tx,
	}
}

// ============================================================
// 新增购物车信息
// ============================================================

// CreateCartItem 创建购物车记录
// 接收值：cartItem - 购物车对象指针
// 返回值：error - 错误信息
func (ci *CartItemRepo) CreateCartItem(cartItem *model.UserCartItem) error {
	return ci.db.Create(cartItem).Error
}

// ============================================================
// 查询购物车信息
// ============================================================

// GetCartItem 查询购物车信息（按购物车ID查）
// 接收值：cartItemId - 购物车唯一标识
// 返回值：
//
//	*response.CartItemListResp - 购物车对象指针
//	error - 购物车不存在返回CartItemNotExist
func (ci *CartItemRepo) GetCartItem(cartItemId int64) (cartItem *response.CartItemListResp, err error) {
	cartItemTable := model.UserCartItem{}.TableName()

	baseQuery := ci.buildCartItemQuery()

	err = baseQuery.Where(cartItemTable+".cart_item_id = ?", cartItemId).
		Scan(&cartItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.CartItemNotExist
		}
		return nil, err
	}
	return cartItem, err
}

// buildCartItemQuery 构建购物车信息联表链式查询（内部方法）
func (ci *CartItemRepo) buildCartItemQuery() *gorm.DB {
	cartItemTable := model.UserCartItem{}.TableName()
	skuTable := model.SysProductSku{}.TableName()
	spuTable := model.SysProductSpu{}.TableName()

	return ci.db.Table(cartItemTable).
		Select(cartItemTable+".*, "+
			skuTable+".sku_id, "+
			skuTable+".sku_name, "+
			skuTable+".spec_values, "+
			skuTable+".sku_image, "+
			skuTable+".sku_status, "+
			skuTable+".price, "+
			skuTable+".stock, "+
			spuTable+".spu_id, "+
			spuTable+".spu_name, "+
			spuTable+".spu_status, "+
			spuTable+".main_image").
		Joins("LEFT JOIN "+skuTable+" ON "+skuTable+".sku_id = "+cartItemTable+".sku_id "+" AND "+skuTable+".is_deleted = ?", false).
		Joins("LEFT JOIN "+spuTable+" ON "+spuTable+".spu_id = "+skuTable+".spu_id"+" AND "+spuTable+".is_deleted = ?", false)
}

// GetCartItemList 查询用户所有购物车列表（默认最近更新购物车排最前）
// 接收值：userId - 用户ID
// 返回值：
//
//	[]response.CartItemListResp - 购物车信息列表
//	error - 列表为空返回CartItemNotExist
func (ci *CartItemRepo) GetCartItemList(userId int64) ([]response.CartItemListResp, error) {
	cartItemTable := model.UserCartItem{}.TableName()
	var cartItems []response.CartItemListResp

	baseQuery := ci.buildCartItemQuery()

	err := baseQuery.Where(cartItemTable+".user_id = ?", userId).
		Order(cartItemTable + ".updated_at DESC").
		Scan(&cartItems).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.CartItemNotExist
		}
		return nil, err
	}
	fmt.Print(cartItems)
	return cartItems, nil
}

// GetCartItemResp 查询用户购物车信息（根据用户ID和所购商品skuId）
// 接收值：
//
//	userId - 用户ID
//	skuId int64
//
// 返回值：
//
//	*response.CartItemListResp - 默认购物车对象指针
//	error - 无默认购物车返回CartItemNotExist
func (ci *CartItemRepo) GetCartItemResp(userId int64, skuId int64) (cartItem *response.CartItemListResp, err error) {
	cartItemTable := model.UserCartItem{}.TableName()

	baseQuery := ci.buildCartItemQuery()

	err = baseQuery.Where(cartItemTable+".user_id = ? AND "+cartItemTable+".sku_id = ?", userId, skuId).
		Scan(&cartItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.CartItemNotExist
		}
		return nil, err
	}
	return cartItem, err
}

// GetCartItemTotal 查询用户当前购物车总数
// 接收值：userId - 用户ID
// 返回值：
//
//	int64 - 当前购物车总数
//	error - 错误信息
func (ci *CartItemRepo) GetCartItemTotal(userId int64) (int64, error) {
	var total int64
	err := ci.db.Model(model.UserCartItem{}).
		Where("user_id = ?", userId).
		Count(&total).Error
	return total, err
}

// GetSelectCartItem 查询用户当前已选中购物车列表
// 接收值：userId - 用户ID
// 返回值：
//
//	[]response.CartItemListResp - 已选中购物车列表
//	error - 错误信息
func (ci *CartItemRepo) GetSelectCartItem(userId int64) ([]response.CartItemListResp, error) {
	cartItemTable := model.UserCartItem{}.TableName()
	var cartItems []response.CartItemListResp

	baseQuery := ci.buildCartItemQuery()

	err := baseQuery.Where(cartItemTable+".user_id = ? AND "+cartItemTable+".is_selected = ?", userId, true).
		Order(cartItemTable + ".updated_at DESC").
		Scan(&cartItems).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.CartItemNotExist
		}
		return nil, err
	}
	return cartItems, nil
}

// ============================================================
// 更新购物车信息
// ============================================================

// UpdateCartItem 更新购物车信息
// 接收值：
//
//	cartItemId - 购物车ID
//	requset.UpdateCartItemReq - 更新参数（选中状态和购买数）
//
// 返回值： error - 错误信息
func (ci *CartItemRepo) UpdateCartItem(cartItemId int64, updateCartItem requset.UpdateCartItemReq) error {
	return ci.db.Model(model.UserCartItem{}).
		Where("cart_item_id = ?", cartItemId).
		Updates(map[string]interface{}{
			"is_selected": updateCartItem.IsSelected,
			"quantity":    updateCartItem.Quantity,
			"updated_at":  time.Now(),
		}).Error
}

// UpdateCartItemQuantity 更新购物车信息
// 接收值：
//
//	cartItemId - 购物车ID
//	quantity - 更新购买数
//
// 返回值： error - 错误信息
func (ci *CartItemRepo) UpdateCartItemQuantity(cartItemId int64, quantity int64) error {
	return ci.db.Model(model.UserCartItem{}).
		Where("cart_item_id = ?", cartItemId).
		Updates(map[string]interface{}{
			"quantity":   quantity,
			"updated_at": time.Now(),
		}).Error
}

// SelectAllCartItem 更新购物车选中状态（全选）
// 接收值：
//
//	cartItemId - 购物车ID
//	isSelected - 选择状态
//
// 返回值：
//
//	int64 - 被影响的购物车数量
//	error - 错误信息
func (ci *CartItemRepo) SelectAllCartItem(userId int64, isSelected bool) (int64, error) {
	var total int64
	err := ci.db.Model(model.UserCartItem{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"is_selected": isSelected,
			"updated_at":  time.Now(),
		})
	if err.Error != nil {
		return 0, err.Error
	}
	total = err.RowsAffected
	return total, nil
}

// ============================================================
// 删除购物车信息
// ============================================================

// DeleteCartItem 删除购物车
// 接收值： cartItemId - 购物车ID
// 返回值： error - 错误信息
func (ci *CartItemRepo) DeleteCartItem(cartItemId int64) error {
	return ci.db.Delete(&model.UserCartItem{}, cartItemId).Error
}
