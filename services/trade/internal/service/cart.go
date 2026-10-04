package service

import (
	"context"
	"errors"
	"fmt"

	"demo-shop/services/trade/internal/model"

	"gorm.io/gorm"
)

// ============================================================
// 购物车领域操作
// ============================================================
//
// 与商品域的边界:购物车只存 sku_id 与数量,所有商品展示字段
// (名称/图/规格/价格/库存)在读写时经 product-service 的 BatchGetSkus 回填。
// 若存快照,商品改价后购物车显示旧价,而用户下单按新价 —— 页面与结算对不上。

// CreateCartItem 加购。
func (c *CartService) CreateCartItem(ctx context.Context, userId, skuId, quantity int64) (*model.CartItemView, error) {
	if err := validateQuantity(quantity); err != nil {
		return nil, err
	}

	// 校验 SKU 存在且可购买 —— 加购时挡一次,结算时还会再挡一次
	// (下单到结算之间商品可能下架,所以两处都要校验,不能只靠这里)
	snap, err := c.getSku(userId, skuId)
	if err != nil {
		return nil, err
	}

	existing, err := c.cartRepo.GetCartItemByUserAndSku(userId, skuId)
	switch {
	case err == nil:
		// 累加数量。上限校验要在累加后判断 ——
		// 只校验单次传入的数量,会漏掉"多次少量加购累计超限"
		newQty := existing.Quantity + quantity
		if newQty > model.CartQuantityMax {
			return nil, model.QuantityMax
		}
		if err := c.cartRepo.AddCartItemQuantity(existing.CartItemId, quantity); err != nil {
			return nil, err
		}
		existing.Quantity = newQty
		return c.toView(existing, snap), nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		// 先校验行数上限(与单体的 100 条一致)
		count, err := c.cartRepo.CountCartItem(userId)
		if err != nil {
			return nil, err
		}
		if count >= model.CartItemMaxCount {
			return nil, model.CartItemMax
		}

		item := &model.UserCartItem{
			UserId:   userId,
			SkuId:    skuId,
			Quantity: quantity,
			// 新加购默认选中:用户刚放进去的东西通常是马上要买的
			IsSelected: true,
		}
		if err := c.cartRepo.CreateCartItem(item); err != nil {
			return nil, err
		}
		return c.toView(item, snap), nil

	default:
		return nil, err
	}
}

// GetCartItem 查某用户购物车中指定 SKU 的行。
func (c *CartService) GetCartItem(ctx context.Context, userId, skuId int64) (*model.CartItemView, error) {
	item, err := c.cartRepo.GetCartItemByUserAndSku(userId, skuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// 回填商品快照。这里不校验可购买
	snapshots, err := c.productRPC.BatchGetSkus([]int64{skuId})
	if err != nil {
		return nil, err
	}
	return c.toView(item, snapshots[skuId]), nil
}

// ListCartItems 购物车列表。
func (c *CartService) ListCartItems(userId int64) ([]*model.CartItemView, error) {
	return c.loadCartItemViews(userId, false)
}

// GetCartPayPreview 结算预览:只统计已选中的行。
func (c *CartService) GetCartPayPreview(userId int64) ([]*model.CartItemView, int64, float64, error) {
	views, err := c.loadCartItemViews(userId, true)
	if err != nil {
		return nil, 0, 0, err
	}

	var totalQty int64
	var totalAmount float64
	for _, v := range views {
		totalQty += v.Quantity
		totalAmount += v.Price * float64(v.Quantity)
	}
	return views, totalQty, totalAmount, nil
}

// UpdateCartItem 改数量 / 改选中态。
func (c *CartService) UpdateCartItem(ctx context.Context, cartItemId, userId, quantity int64, isSelected bool) (*model.CartItemView, error) {
	item, err := c.ownedCartItem(cartItemId, userId)
	if err != nil {
		return nil, err
	}
	// quantity <= 0 时沿用原值:接口同时承担"只切换选中态"的用途,
	// 那种调用不会传数量
	if quantity > 0 {
		if err := validateQuantity(quantity); err != nil {
			return nil, err
		}
		item.Quantity = quantity
	}
	if err := c.cartRepo.UpdateCartItem(cartItemId, item.Quantity, isSelected); err != nil {
		return nil, err
	}
	item.IsSelected = isSelected

	// 回填时按后端最新价重算:用户可能刚改完数量就看到金额变化
	snap, err := c.getSku(userId, item.SkuId)
	if err != nil {
		return nil, err
	}
	return c.toView(item, snap), nil
}

// DeleteCartItem 删除购物车行
func (c *CartService) DeleteCartItem(cartItemId, userId int64) error {
	if _, err := c.ownedCartItem(cartItemId, userId); err != nil {
		return err
	}
	return c.cartRepo.DeleteCartItem(cartItemId)
}

// SelectAllCartItems 全选 / 取消全选,返回受影响行数
func (c *CartService) SelectAllCartItems(userId int64, isSelected bool) (int64, error) {
	return c.cartRepo.UpdateSelectAll(userId, isSelected)
}

// GetCartItemCount 购物车行数(商城顶栏角标用)
func (c *CartService) GetCartItemCount(userId int64) (int64, error) {
	return c.cartRepo.CountCartItem(userId)
}

// ============================================================
// 内部辅助
// ============================================================

// loadCartItemViews 取购物车行并回填商品快照。
func (c *CartService) loadCartItemViews(userId int64, selectedOnly bool) ([]*model.CartItemView, error) {
	var (
		items []*model.UserCartItem
		err   error
	)
	if selectedOnly {
		items, err = c.cartRepo.GetSelectedCartItemList(userId)
	} else {
		items, err = c.cartRepo.GetCartItemList(userId)
	}
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		// 空车是正常状态,不是错误 —— 前端展示"购物车是空的"
		return []*model.CartItemView{}, nil
	}

	skuIds := make([]int64, 0, len(items))
	for _, it := range items {
		skuIds = append(skuIds, it.SkuId)
	}
	snapshots, err := c.productRPC.BatchGetSkus(skuIds)
	if err != nil {
		return nil, err
	}

	views := make([]*model.CartItemView, 0, len(items))
	for _, it := range items {
		// 商品侧查不到(已物理删除)也要返回这一行,只是 Available=false ——
		// 静默丢弃会让用户以为商品凭空消失
		views = append(views, c.toView(it, snapshots[it.SkuId]))
	}
	return views, nil
}

// ownedCartItem 取购物车行并校验归属。
func (c *CartService) ownedCartItem(cartItemId, userId int64) (*model.UserCartItem, error) {
	item, err := c.cartRepo.GetCartItemById(cartItemId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.CartItemNotExist
		}
		return nil, err
	}
	if item.UserId != userId {
		return nil, model.NotAuthority
	}
	return item, nil
}

// getSku 取单个 SKU 快照并校验可购买。
func (c *CartService) getSku(userId, skuId int64) (*SkuSnapshot, error) {
	snapshots, err := c.productRPC.BatchGetSkus([]int64{skuId})
	if err != nil {
		return nil, err
	}
	snap, ok := snapshots[skuId]
	if !ok || snap == nil {
		// 查不到 = SKU 不存在或已物理删除
		return nil, model.ErrSkuNotExist
	}
	if !snap.Available {
		// 存在但不可购买(下架/禁用/SPU 未发布)
		return nil, model.ErrSpuDisabled
	}
	return snap, nil
}

// toView 购物车行 + SKU 快照 → 视图。
func (c *CartService) toView(item *model.UserCartItem, snap *SkuSnapshot) *model.CartItemView {
	view := &model.CartItemView{UserCartItem: *item}
	if snap == nil {
		// 商品侧查不到(已物理删除):仍返回行本身,标为不可购买
		view.Available = false
		view.UnavailableReason = model.ProductWithdraw
		return view
	}
	view.SpuId = snap.SpuId
	view.SpuName = snap.SpuName
	view.SkuName = snap.SkuName
	view.SpecValues = snap.SpecValues
	view.MainImage = snap.MainImage
	view.SkuImage = snap.SkuImage
	view.Price = snap.Price
	view.Stock = snap.Stock

	// 可购买性判断与原因文案都由服务端给:原因里带库存数字
	// ("仅剩 2 件"),前端拼还得再取一次库存
	switch {
	case !snap.Available:
		view.Available = false
		view.UnavailableReason = model.ProductWithdraw
	case snap.Stock < item.Quantity:
		// 有货但不够买这么多 —— 与"已下架"是两回事,分开给原因
		view.Available = false
		view.UnavailableReason = fmt.Sprintf("库存不足,仅剩%d件", snap.Stock)
	default:
		view.Available = true
	}
	return view
}

// validateQuantity 数量校验。
func validateQuantity(quantity int64) error {
	if quantity < 1 || quantity > model.CartQuantityMax {
		return model.QuantityMax
	}
	return nil
}
