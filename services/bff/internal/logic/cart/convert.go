package cart

import (
	"encoding/json"

	"demo-shop/services/bff/internal/types"

	v1_tradev1 "demo-shop/api/gen/trade/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// proto CartItem → types.CartItem
// ============================================================
//
// 单体的转换在 tradeclient.toModelCartItem 里,做了三类改动:
//
//	改名  total_price → subtotal
//	      available   → is_available
//	类型  spec_values 从 **JSON 文本字符串** 变成 **对象**
//	隐藏  user_id / created_at / updated_at / sku_status / spu_status
//	      (单体在 response.CartItemListResp 上标了 `json:"-"`)
//
// 这三类都必须照做 —— 前端解的是 subtotal / is_available / 对象形态的
// spec_values。照 proto 的字段名直传会让前端拿到 undefined,**而且不报错**。

func toCartItem(p *v1_tradev1.CartItem) types.CartItem {
	if p == nil {
		return types.CartItem{}
	}
	return types.CartItem{
		CartItemId: p.GetCartItemId(),
		SkuId:      p.GetSkuId(),
		SpuId:      p.GetSpuId(),
		SpuName:    p.GetSpuName(),
		MainImage:  p.GetMainImage(),
		SkuName:    p.GetSkuName(),
		SpecValues: parseSpecValues(p.GetSpecValues()),
		SkuImage:   p.GetSkuImage(),
		Price:      p.GetPrice(),
		Stock:      p.GetStock(),
		Quantity:   p.GetQuantity(),
		IsSelected: p.GetIsSelected(),
		// 改名:proto 的 total_price → HTTP 的 subtotal
		Subtotal: p.GetTotalPrice(),
		// 改名:proto 的 available → HTTP 的 is_available
		IsAvailable:       p.GetAvailable(),
		UnavailableReason: p.GetUnavailableReason(),
	}
}

func toCartItems(ps []*v1_tradev1.CartItem) []types.CartItem {
	// make(..., 0, len) 而不是 var x []T —— 后者无数据时是 nil,
	// 序列化成 null 而非 [],而前端有 list.length / .map() 这类写法。
	out := make([]types.CartItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, toCartItem(p))
	}
	return out
}

// parseSpecValues 把 proto 的 JSON 文本解析成对象。
//
// ============================================================
// 为什么解析失败时返回空对象而不是报错
// ============================================================
//
// 这是一个**刻意的静默降级**,与项目里"未知错误必须落到 500 让运维
// 看见"的原则相反。理由:
//
//	spec_values 只是规格展示(如 {"颜色":"红","尺码":"XL"}),
//	它解析失败说明 trade-service 那条数据有问题 —— 但那不该让
//	**整个购物车列表**失败。用户看到的应当是"这一行没显示规格",
//	而不是"购物车打不开"。
//
// 空对象({})而不是 nil:前端可能在渲染时做 spec_values.颜色 这类
// 取值,nil 会抛错而 {} 安全。而 JSON 序列化时 {} 也比 null 更接近
// "有规格但为空"的语义。
//
// **代价**:这类问题不会以错误形式暴露,只会在日志里。故这里必须
// 记日志 —— 否则 trade-service 返回坏数据时两边都静默。
//
// 单体那边是 jsonTextToMap,行为相同(失败返回空 map)。
func parseSpecValues(s string) interface{} {
	// 空串是**正常情况**(proto 注释:available=true 时为空串之外,
	// 商品侧回填字段也可能为空)。走解析会得到一个 "unexpected end of
	// JSON input" 错误并记一条无意义的日志,故先短路。
	if s == "" {
		return map[string]interface{}{}
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		// 用 logx 而不是把错误往上抛 —— 见上方说明。
		// **不打印原始值**:spec_values 里理论上没有敏感信息,
		// 但它是商品侧完全可控的输入,原样打进日志会让日志格式
		// 被外部数据影响(如超长字符串、换行)。只打长度。
		logx.Errorf("cart: spec_values 不是合法 JSON(len=%d): %v", len(s), err)
		return map[string]interface{}{}
	}
	if m == nil {
		// `null` 是合法 JSON,Unmarshal 进 map 得到 nil。
		// 归一到 {} 保持一致。
		return map[string]interface{}{}
	}
	return m
}
