package server

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	paymentservicelogic "demo-shop/services/trade/internal/logic/paymentservice"
)

// ============================================================
// PaymentService 的 gRPC 适配层(纯转发)
// ============================================================

func (s *PaymentServiceServer) CreatePayment(ctx context.Context, in *v1_tradev1.CreatePaymentReq) (*v1_tradev1.CreatePaymentResp, error) {
	return paymentservicelogic.NewCreatePaymentLogic(ctx, s.svcCtx).CreatePayment(in)
}

func (s *PaymentServiceServer) GetPayment(ctx context.Context, in *v1_tradev1.GetPaymentReq) (*v1_tradev1.GetPaymentResp, error) {
	return paymentservicelogic.NewGetPaymentLogic(ctx, s.svcCtx).GetPayment(in)
}

func (s *PaymentServiceServer) HandlePaymentCallback(ctx context.Context, in *v1_tradev1.HandlePaymentCallbackReq) (*v1_tradev1.HandlePaymentCallbackResp, error) {
	return paymentservicelogic.NewHandlePaymentCallbackLogic(ctx, s.svcCtx).HandlePaymentCallback(in)
}

func (s *PaymentServiceServer) ListPayments(ctx context.Context, in *v1_tradev1.ListPaymentsReq) (*v1_tradev1.ListPaymentsResp, error) {
	return paymentservicelogic.NewListPaymentsLogic(ctx, s.svcCtx).ListPayments(in)
}
