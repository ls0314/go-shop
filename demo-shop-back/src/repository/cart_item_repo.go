package repository

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
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
func NewCartItemRepo(conn *gorm.DB) *CartItemRepo {
	return &CartItemRepo{db: conn}
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

// GetCartItem 查询购物车信息（按购物车ID查）。
// 只返回 user_cart_item 自身的列;商品侧字段(sku_name/price/spu_name/main_image…)
// 由 service 层经 product-service RPC 回填 —— 商品表的所有权不在本库,不能再 JOIN。
// 接收值：cartItemId - 购物车唯一标识
// 返回值：
//
//	*response.CartItemListResp - 购物车对象指针
//	error - 购物车不存在返回CartItemNotExist
func (ci *CartItemRepo) GetCartItem(cartItemId int64) (cartItem *response.CartItemListResp, err error) {
	cartItemTable := model.UserCartItem{}.TableName()

	err = ci.buildCartItemQuery().
		Where(cartItemTable+".cart_item_id = ?", cartItemId).
		Scan(&cartItem).Error
	if err != nil {
		return nil, err
	}
	// Scan 对空结果不返回 ErrRecordNotFound,只留零值 —— 必须显式判主键,
	// 否则调用方会拿到一个 cart_item_id=0 的"存在"记录。
	if cartItem == nil || cartItem.CartItemId == 0 {
		return nil, model.CartItemNotExist
	}
	return cartItem, err
}

// buildCartItemQuery 构建购物车基础查询（内部方法）。
//
// 只查 user_cart_item 单表:user_cart_item 是购物车域的表,归本服务;
// sys_product_sku / sys_product_spu 归 product-service,跨库 JOIN 会破坏分库边界
// (且两边已分库,JOIN 根本跑不通)。商品字段的填充见 service 层的 fillCartItemProducts。
func (ci *CartItemRepo) buildCartItemQuery() *gorm.DB {
	cartItemTable := model.UserCartItem{}.TableName()

	return ci.db.Table(cartItemTable).
		Select(cartItemTable + ".*")
}

// GetCartItemList 查询用户所有购物车列表（默认最近更新购物车排最前）
// 接收值：userId - 用户ID
// 返回值：
//
//	[]response.CartItemListResp - 购物车信息列表(商品字段待 service 层回填)
//	error - 错误信息
func (ci *CartItemRepo) GetCartItemList(userId int64) ([]response.CartItemListResp, error) {
	cartItemTable := model.UserCartItem{}.TableName()
	var cartItems []response.CartItemListResp

	err := ci.buildCartItemQuery().
		Where(cartItemTable+".user_id = ?", userId).
		Order(cartItemTable + ".updated_at DESC").
		Scan(&cartItems).Error
	if err != nil {
		return nil, err
	}
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

	err = ci.buildCartItemQuery().
		Where(cartItemTable+".user_id = ? AND "+cartItemTable+".sku_id = ?", userId, skuId).
		Scan(&cartItem).Error
	if err != nil {
		return nil, err
	}
	if cartItem == nil || cartItem.CartItemId == 0 {
		return nil, model.CartItemNotExist
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
//	[]response.CartItemListResp - 已选中购物车列表(商品字段待 service 层回填)
//	error - 错误信息
func (ci *CartItemRepo) GetSelectCartItem(userId int64) ([]response.CartItemListResp, error) {
	cartItemTable := model.UserCartItem{}.TableName()
	var cartItems []response.CartItemListResp

	err := ci.buildCartItemQuery().
		Where(cartItemTable+".user_id = ? AND "+cartItemTable+".is_selected = ?", userId, true).
		Order(cartItemTable + ".updated_at DESC").
		Scan(&cartItems).Error
	if err != nil {
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
