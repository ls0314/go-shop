package server

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"
)

// PaymentServiceServer 支付的 gRPC 适配层。
// 只做"proto 入参 → logic → proto 出参"的搬运,不含业务规则。
type PaymentServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_tradev1.UnimplementedPaymentServiceServer
}

func NewPaymentServiceServer(svcCtx *svc.ServiceContext) *PaymentServiceServer {
	return &PaymentServiceServer{
		svcCtx: svcCtx,
	}
}
