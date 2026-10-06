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
		Username:  r.Username,
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
	// OrderNo 来自联表;Username 来自流水表自己的快照列(ToProtoPayment 已填)
	p.OrderNo = v.OrderNo
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
