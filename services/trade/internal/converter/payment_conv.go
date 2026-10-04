package converter

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/model"
)

// ============================================================
// 支付:proto ↔ 领域对象
// ============================================================

// ToProtoPayment 支付流水 → proto。
func ToProtoPayment(r *model.UserPaymentRecord) *v1_tradev1.Payment {
	if r == nil {
		return nil
	}
	return &v1_tradev1.Payment{
		PaymentId: r.PaymentId,
		PayNo:     r.PayNo,
		OrderId:   r.OrderId,
		PayMethod: r.PayMethod,
		PayAmount: r.PayAmount,
		PayStatus: r.PayStatus,
		PayTime:   timestampOrNil(r.PayTime),
		TradeNo:   r.TradeNo,
		CreatedAt: timestampOrNil(r.CreatedAt),
		UpdatedAt: timestampOrNil(r.UpdatedAt),
	}
}

// ToProtoPaymentView 联表视图 → proto(带订单号,列表页用)
func ToProtoPaymentView(v *model.PaymentView) *v1_tradev1.Payment {
	if v == nil {
		return nil
	}
	p := ToProtoPayment(&v.UserPaymentRecord)
	p.OrderNo = v.OrderNo
	// Username 由调用方(user 域)回填:用户名在 user_db,跨库取不到。
	// 契约里为它留了字段,本服务不假装能填
	return p
}

// ToProtoPaymentViews 批量转换
func ToProtoPaymentViews(items []*model.PaymentView) []*v1_tradev1.Payment {
	out := make([]*v1_tradev1.Payment, 0, len(items))
	for _, v := range items {
		out = append(out, ToProtoPaymentView(v))
	}
	return out
}
