package server

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/svc"
)

// CategoryServiceServer 类目服务的 gRPC 适配层。
// 只做"proto 入参 → logic → proto 出参"的搬运,不含业务规则;
// 各 RPC 方法集中在同包的 category.go 里。
type CategoryServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_productv1.UnimplementedCategoryServiceServer
}

func NewCategoryServiceServer(svcCtx *svc.ServiceContext) *CategoryServiceServer {
	return &CategoryServiceServer{
		svcCtx: svcCtx,
	}
}
