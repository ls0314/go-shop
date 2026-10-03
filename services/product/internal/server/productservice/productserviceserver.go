package server

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/svc"
)

// ProductServiceServer 商品服务的 gRPC 适配层。
// 只做"proto 入参 → logic → proto 出参"的搬运,不含业务规则;
// 各 RPC 方法集中在同包的 product.go 里。
type ProductServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_productv1.UnimplementedProductServiceServer
}

func NewProductServiceServer(svcCtx *svc.ServiceContext) *ProductServiceServer {
	return &ProductServiceServer{
		svcCtx: svcCtx,
	}
}
