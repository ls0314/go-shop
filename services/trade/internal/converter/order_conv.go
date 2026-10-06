package converter

import (
	"encoding/json"
	"time"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/dto/req"
	"demo-shop/services/trade/internal/dto/resp"
	"demo-shop/services/trade/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
)

// ============================================================
// proto ↔ 领域对象
// ============================================================

// FromProtoCreateOrder 下单入参:proto → 领域 req
func FromProtoCreateOrder(p *v1_tradev1.CreateOrderReq) *req.CreateOrderReq {
	if p == nil {
		return &req.CreateOrderReq{}
	}
	out := &req.CreateOrderReq{
		UserId:        p.GetUserId(),
		UserName:      p.GetUsername(),
		IdempotentKey: p.GetIdempotentKey(),
		BuyerRemark:   p.GetBuyerRemark(),
		UserCouponId:  p.GetUserCouponId(),
	}
	// 地址快照是**一次请求的形状**,不是实体 —— 故在此从 proto 拷到 model.AddressSnap
	// (后者同时是 user_order_master.address_snapshot 的落库形状)
	if a := p.GetAddress(); a != nil {
		out.AddressSnapshot = model.AddressSnap{
			ReceiverName:  a.GetReceiverName(),
			ReceiverPhone: a.GetReceiverPhone(),
			Province:      a.GetProvince(),
			City:          a.GetCity(),
			District:      a.GetDistrict(),
			DetailAddress: a.GetDetailAddress(),
			PostalCode:    a.GetPostalCode(),
		}
	}
	return out
}

// FromProtoCancelOrder 取消入参:proto → 领域 req
func FromProtoCancelOrder(p *v1_tradev1.CancelOrderReq) *req.CancelOrderReq {
	if p == nil {
		return &req.CancelOrderReq{}
	}
	return &req.CancelOrderReq{
		OrderId:  p.GetOrderId(),
		UserId:   p.GetUserId(),
		Operator: p.GetOperator(),
	}
}

// FromProtoListUserOrders 用户端列表入参
func FromProtoListUserOrders(p *v1_tradev1.ListUserOrdersReq) *req.ListUserOrdersReq {
	if p == nil {
		return &req.ListUserOrdersReq{}
	}
	return &req.ListUserOrdersReq{
		Page:        int(p.GetPage()),
		PageSize:    int(p.GetPageSize()),
		OrderStatus: p.GetOrderStatus(),
	}
}

// FromProtoListOrders 管理端列表入参
func FromProtoListOrders(p *v1_tradev1.ListOrdersReq) *req.ListOrdersReq {
	if p == nil {
		return &req.ListOrdersReq{}
	}
	out := &req.ListOrdersReq{
		Page:        int(p.GetPage()),
		PageSize:    int(p.GetPageSize()),
		OrderStatus: p.GetOrderStatus(),
		OrderNo:     p.GetOrderNo(),
	}
	// 时间区间是可选筛选项:proto 用 nil 表达"没传",
	// 领域层用 *time.Time 表达同一件事,故不能直接 AsTime() ——
	// 那样零值会变成一个真实的公元 1 年
	if p.StartTime != nil {
		t := p.StartTime.AsTime()
		out.StartTime = &t
	}
	if p.EndTime != nil {
		t := p.EndTime.AsTime()
		out.EndTime = &t
	}
	return out
}

// FromProtoShipOrder 发货入参
func FromProtoShipOrder(p *v1_tradev1.ShipOrderReq) *req.ShipOrderReq {
	if p == nil {
		return &req.ShipOrderReq{}
	}
	return &req.ShipOrderReq{
		OrderId:        p.GetOrderId(),
		UserName:       p.GetUsername(),
		ExpressCompany: p.GetExpressCompany(),
		TrackingNo:     p.GetTrackingNo(),
	}
}

// ============================================================
// 领域对象 → proto
// ============================================================

// ToProtoCreateOrderResp 下单结果 → proto
func ToProtoCreateOrderResp(r *resp.CreateOrderResp) *v1_tradev1.CreateOrderResp {
	if r == nil {
		return &v1_tradev1.CreateOrderResp{}
	}
	return &v1_tradev1.CreateOrderResp{
		OrderId:     r.OrderId,
		OrderNo:     r.OrderNo,
		TotalAmount: r.TotalAmount,
		PayAmount:   r.PayAmount,
		OrderStatus: r.OrderStatus,
		PayExpireAt: timestampOrNil(r.PayExpireAt),
		CreatedAt:   timestampOrNil(r.CreatedAt),
	}
}

// ToProtoOrder 订单实体 → proto。
func ToProtoOrder(o *model.UserOrder) *v1_tradev1.Order {
	if o == nil {
		return nil
	}
	return &v1_tradev1.Order{
		OrderId:     o.OrderId,
		OrderNo:     o.OrderNo,
		UserId:      o.UserId,
		Username:    o.Username,
		OrderStatus: o.OrderStatus,
		TotalAmount: o.TotalAmount,
		PayAmount:   o.PayAmount,
		PayMethod:   o.PayMethod,
		PayTime:     timestampOrNil(o.PayTime),
		// AddressSnapshot 在库里是 jsonb、在 proto 里是 JSON 文本 —— 直接转
		AddressSnapshot: string(o.AddressSnapshot),
		BuyerRemark:     o.BuyerRemark,
		DetailCount:     o.DetailCount,
		FirstImage:      o.FirstImage,
		ExpireAt:        timestampOrNil(o.ExpireAt),
		IdempotentKey:   o.IdempotentKey,
		CreatedAt:       timestampOrNil(o.CreatedAt),
		UpdatedAt:       timestampOrNil(o.UpdatedAt),
	}
}

// ToProtoOrderList 订单列表 → proto
func ToProtoOrderList(items []*model.UserOrder) []*v1_tradev1.Order {
	out := make([]*v1_tradev1.Order, 0, len(items))
	for _, o := range items {
		out = append(out, ToProtoOrder(o))
	}
	return out
}

// ToProtoOrderDetail 明细实体 → proto
func ToProtoOrderDetail(d *model.UserOrderDetail) *v1_tradev1.OrderDetail {
	if d == nil {
		return nil
	}
	return &v1_tradev1.OrderDetail{
		DetailId: d.DetailId,
		OrderId:  d.OrderId,
		SkuId:    d.SkuId,
		SpuName:  d.SpuName,
		SkuName:  d.SkuName,
		// SpecValues 在库里是 jsonb、在 proto 里是 JSON 文本
		SpecValues: jsonMapToText(d.SpecValues),
		MainImage:  d.MainImage,
		Quantity:   d.Quantity,
		UnitPrice:  d.UnitPrice,
		TotalPrice: d.TotalPrice,
	}
}

// ToProtoOrderDetails 明细列表 → proto
func ToProtoOrderDetails(items []*model.UserOrderDetail) []*v1_tradev1.OrderDetail {
	out := make([]*v1_tradev1.OrderDetail, 0, len(items))
	for _, d := range items {
		out = append(out, ToProtoOrderDetail(d))
	}
	return out
}

// ToProtoOrderLog 日志实体 → proto
func ToProtoOrderLog(l *model.UserOrderLog) *v1_tradev1.OrderLog {
	if l == nil {
		return nil
	}
	return &v1_tradev1.OrderLog{
		LogId:       l.LogId,
		OrderId:     l.OrderId,
		OrderStatus: l.OrderStatus,
		Action:      l.Action,
		Operator:    l.Operator,
		Detail:      l.Detail,
		CreatedAt:   timestampOrNil(l.CreatedAt),
	}
}

// ToProtoOrderLogs 日志列表 → proto
func ToProtoOrderLogs(items []*model.UserOrderLog) []*v1_tradev1.OrderLog {
	out := make([]*v1_tradev1.OrderLog, 0, len(items))
	for _, l := range items {
		out = append(out, ToProtoOrderLog(l))
	}
	return out
}

// timestampOrNil 零值时间转 nil。
//
// 为什么不能直接 timestamppb.New(t):零值时间会变成公元 1 年,
// 前端渲染出来是 "0001-01-01",而语义上它是"没有这个时间"。
func timestampOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// jsonMapToText 把 jsonb 列的值转成 proto 里的 JSON 文本。
//
// 库里的 spec_values 是 jsonb、proto 里是 string —— 两个形状不一致是刻意的:
// proto 用文本是为了让契约不依赖某个具体语言的 JSON 类型。
// 序列化失败(理论上不会,jsonb 出来的一定是合法 JSON)返回空对象,
// 不让展示用的规格字段阻断订单详情接口。
func jsonMapToText(m datatypes.JSONMap) string {
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}
