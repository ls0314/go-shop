package server

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"
)

// OrderServiceServer 订单的 gRPC 适配层。
type OrderServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_tradev1.UnimplementedOrderServiceServer
}

func NewOrderServiceServer(svcCtx *svc.ServiceContext) *OrderServiceServer {
	return &OrderServiceServer{
		svcCtx: svcCtx,
	}
}
