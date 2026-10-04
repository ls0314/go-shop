package server

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"
)

// CartServiceServer 购物车的 gRPC 适配层。
// 只做"proto 入参 → logic → proto 出参"的搬运,不含业务规则。
type CartServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_tradev1.UnimplementedCartServiceServer
}

func NewCartServiceServer(svcCtx *svc.ServiceContext) *CartServiceServer {
	return &CartServiceServer{
		svcCtx: svcCtx,
	}
}
