package repository

import (
	"demo-shop/services/trade/internal/model"

	"gorm.io/gorm"
)

// CartItemRepo 购物车表数据层实例
type CartItemRepo struct {
	DB *gorm.DB
}

// NewCartItemRepo 创建购物车表数据层实例
func NewCartItemRepo(conn *gorm.DB) *CartItemRepo {
	return &CartItemRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (c *CartItemRepo) WithTx(tx *gorm.DB) *CartItemRepo {
	return &CartItemRepo{DB: tx}
}

// CreateCartItem 新增购物车行,回填自增主键
func (c *CartItemRepo) CreateCartItem(item *model.UserCartItem) error {
	return c.DB.Create(item).Error
}

// GetCartItemById 根据ID查询购物车行。
func (c *CartItemRepo) GetCartItemById(cartItemId int64) (*model.UserCartItem, error) {
	var item model.UserCartItem
	if err := c.DB.Where("cart_item_id = ?", cartItemId).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetCartItemByUserAndSku 按 用户 + SKU 查购物车行。
func (c *CartItemRepo) GetCartItemByUserAndSku(userId, skuId int64) (*model.UserCartItem, error) {
	var item model.UserCartItem
	err := c.DB.Where("user_id = ? AND sku_id = ?", userId, skuId).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetCartItemList 查某用户全部购物车行(不分页)。
func (c *CartItemRepo) GetCartItemList(userId int64) ([]*model.UserCartItem, error) {
	var list []*model.UserCartItem
	err := c.DB.Where("user_id = ?", userId).
		Order("updated_at DESC").
		Find(&list).Error
	return list, err
}

// GetSelectedCartItemList 查某用户**已选中**的行(结算用)
func (c *CartItemRepo) GetSelectedCartItemList(userId int64) ([]*model.UserCartItem, error) {
	var list []*model.UserCartItem
	err := c.DB.Where("user_id = ? AND is_selected = ?", userId, true).
		Order("updated_at DESC").
		Find(&list).Error
	return list, err
}

// CountCartItem 统计购物车行数(行数上限校验与顶栏角标共用)
func (c *CartItemRepo) CountCartItem(userId int64) (int64, error) {
	var count int64
	err := c.DB.Model(&model.UserCartItem{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

// UpdateCartItem 更新数量与选中态(局部更新,只改传入的两列)
func (c *CartItemRepo) UpdateCartItem(cartItemId, quantity int64, isSelected bool) error {
	return c.DB.Model(&model.UserCartItem{}).
		Where("cart_item_id = ?", cartItemId).
		Updates(map[string]interface{}{
			"quantity":    quantity,
			"is_selected": isSelected,
		}).Error
}

// AddCartItemQuantity 在现有数量上累加(加购已存在的商品时用)。
func (c *CartItemRepo) AddCartItemQuantity(cartItemId, delta int64) error {
	return c.DB.Model(&model.UserCartItem{}).
		Where("cart_item_id = ?", cartItemId).
		Update("quantity", gorm.Expr("quantity + ?", delta)).Error
}

// DeleteCartItem 删除购物车行(下单成功后清理,或用户主动删除)
func (c *CartItemRepo) DeleteCartItem(cartItemId int64) error {
	return c.DB.Where("cart_item_id = ?", cartItemId).Delete(&model.UserCartItem{}).Error
}

// UpdateSelectAll 全选/全不选。
func (c *CartItemRepo) UpdateSelectAll(userId int64, isSelected bool) (int64, error) {
	res := c.DB.Model(&model.UserCartItem{}).
		Where("user_id = ?", userId).
		Update("is_selected", isSelected)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
