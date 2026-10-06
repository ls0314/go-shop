package order

import (
	"encoding/json"

	"demo-shop/services/bff/internal/types"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// proto OrderDetail → types.OrderDetail
// ============================================================
//
// 单体在 response.UserGetOrderDetail 上做的改动很少(只隐藏了
// order_id),字段名与 proto 一致:detail_id / sku_id / spu_name /
// sku_name / spec_values / main_image / quantity / unit_price /
// total_price。
//
// **注意这里的 spec_values 是 string 而不是对象** —— 与购物车不同!
//
//	购物车 CartItemListResp.SpecValues 是 datatypes.JSONMap → **对象**
//	订单   UserGetOrderDetail.SpecValues 是 string          → **字符串**
//
// 所以订单详情里前端拿到的是字符串("红色:XL"),而购物车里是对象。
// **这是既有的不一致,不要统一** —— 统一任何一边都会破坏那一侧的
// 前端渲染。已核实单体两侧的声明确实不同。
func toOrderDetail(p *v1_tradev1.OrderDetail) types.OrderDetail {
	return types.OrderDetail{
		DetailId:   p.GetDetailId(),
		SkuId:      p.GetSkuId(),
		SpuName:    p.GetSpuName(),
		SkuName:    p.GetSkuName(),
		SpecValues: p.GetSpecValues(),
		MainImage:  p.GetMainImage(),
		Quantity:   p.GetQuantity(),
		UnitPrice:  p.GetUnitPrice(),
		TotalPrice: p.GetTotalPrice(),
	}
}

func toOrderDetails(ps []*v1_tradev1.OrderDetail) []types.OrderDetail {
	out := make([]types.OrderDetail, 0, len(ps))
	for _, p := range ps {
		out = append(out, toOrderDetail(p))
	}
	return out
}

// ============================================================
// proto OrderLog → types.OrderLog
// ============================================================

func toOrderLog(p *v1_tradev1.OrderLog) types.OrderLog {
	return types.OrderLog{
		LogId:       p.GetLogId(),
		OrderId:     p.GetOrderId(),
		OrderStatus: p.GetOrderStatus(),
		Action:      p.GetAction(),
		Operator:    p.GetOperator(),
		Detail:      p.GetDetail(),
		CreatedAt:   formatTimestamp(p.GetCreatedAt()),
	}
}

func toOrderLogs(ps []*v1_tradev1.OrderLog) []types.OrderLog {
	out := make([]types.OrderLog, 0, len(ps))
	for _, p := range ps {
		out = append(out, toOrderLog(p))
	}
	return out
}

// ============================================================
// address_snapshot:string(proto)→ 对象(HTTP)
// ============================================================
//
// proto 的 Order.address_snapshot 是**字符串**,而 HTTP 契约里
// (单体的 datatypes.JSON)是**对象**:
//
//	{ "receiver_name": "...", "receiver_phone": "...", ... }
//
// 故 OrderRPC 的响应里那个字段需要解析。这与 spec_values 是同一类
// 问题,处理方式也一致(解析失败返回空对象 + 记日志)。

// parseSnapshot 把 proto 的 address_snapshot 字符串解析成对象。
//
// 与 cart 的 parseSpecValues 同一策略:解析失败返回**空对象**而不是
// 报错 —— 地址快照是订单详情的展示字段,它坏了不该让整个详情接口
// 失败,而空对象前端能容错(显示为空)。
//
// 同样记日志,否则 trade-service 返回坏数据时两边都静默。
//
// **注意这个函数只用于"读"路径**(列表与详情)。写路径
// (CreateOrder)传的是 proto 的 AddressSnapshot 消息,不经过这里。
func parseSnapshot(s string) interface{} {
	if s == "" {
		return map[string]interface{}{}
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		logx.Errorf("order: address_snapshot 不是合法 JSON(len=%d): %v", len(s), err)
		return map[string]interface{}{}
	}
	if m == nil {
		return map[string]interface{}{}
	}
	return m
}

// ============================================================
// proto AddressSnapshot → 传给 trade 的快照
// ============================================================

// toTradeSnapshot 把 user-service 的快照转成 trade 要的形状。
//
// 两个 message 的字段名完全一致(receiver_name / receiver_phone /
// province / city / district / detail_address / postal_code),
// 故是一一映射。
//
// **为什么不让 trade 自己拿 address_id 去查**:地址表在 user_db,
// trade 跨库读不到;而且订单存的是**下单那一刻**的地址快照,
// 由调用方从权威源取一次再传,语义最贴切。
// (这段理由来自单体 OrderHandler.CreateOrder 的注释。)
func toTradeSnapshot(s *v1_userv1.AddressSnapshot) *v1_tradev1.AddressSnapshot {
	if s == nil {
		return nil
	}
	return &v1_tradev1.AddressSnapshot{
		ReceiverName:  s.GetReceiverName(),
		ReceiverPhone: s.GetReceiverPhone(),
		Province:      s.GetProvince(),
		City:          s.GetCity(),
		District:      s.GetDistrict(),
		DetailAddress: s.GetDetailAddress(),
		PostalCode:    s.GetPostalCode(),
	}
}
