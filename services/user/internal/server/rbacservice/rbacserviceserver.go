package server

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"
)

// RBACServiceServer 平台 RBAC 管理服务的 gRPC 适配层。
// 各域方法分散在同包的 permission.go / role.go 等文件里。
type RBACServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_userv1.UnimplementedRBACServiceServer
}

func NewRBACServiceServer(svcCtx *svc.ServiceContext) *RBACServiceServer {
	return &RBACServiceServer{
		svcCtx: svcCtx,
	}
}
