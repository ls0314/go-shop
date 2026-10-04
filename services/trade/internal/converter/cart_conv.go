package converter

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/model"
)

// ============================================================
// 购物车:proto ↔ 领域对象
// ============================================================

// ToProtoCartItem 购物车视图 → proto。
//
// 商品侧字段(名称/图/规格/价格/库存)是**回填出来的**,不落库 ——
// 故这里直接从视图拷,而不是从实体的列读。
func ToProtoCartItem(v *model.CartItemView) *v1_tradev1.CartItem {
	if v == nil {
		return nil
	}
	return &v1_tradev1.CartItem{
		CartItemId: v.CartItemId,
		UserId:     v.UserId,
		SkuId:      v.SkuId,
		Quantity:   v.Quantity,
		IsSelected: v.IsSelected,
		CreatedAt:  timestampOrNil(v.CreatedAt),
		UpdatedAt:  timestampOrNil(v.UpdatedAt),

		SpuId:      v.SpuId,
		SpuName:    v.SpuName,
		SkuName:    v.SkuName,
		SpecValues: v.SpecValues,
		MainImage:  v.MainImage,
		Price:      v.Price,
		Stock:      v.Stock,
		SkuImage:   v.SkuImage,
		Available:  v.Available,
		// 不可购买的原因由服务端算(带库存数字),前端直接展示
		UnavailableReason: v.UnavailableReason,
		// 小计由服务端算:前端不该自己乘 —— 两处算法漂移时,
		// 页面显示与结算金额会对不上,而用户只信页面上看到的
		TotalPrice: v.Price * float64(v.Quantity),
	}
}

// ToProtoCartItems 批量转换
func ToProtoCartItems(items []*model.CartItemView) []*v1_tradev1.CartItem {
	out := make([]*v1_tradev1.CartItem, 0, len(items))
	for _, v := range items {
		out = append(out, ToProtoCartItem(v))
	}
	return out
}

// ToProtoCartPayPreview 结算预览 → proto
func ToProtoCartPayPreview(items []*model.CartItemView, totalQty int64, totalAmount float64) *v1_tradev1.CartPayPreview {
	return &v1_tradev1.CartPayPreview{
		SelectedItems: ToProtoCartItems(items),
		TotalQuantity: totalQty,
		TotalAmount:   totalAmount,
	}
}
