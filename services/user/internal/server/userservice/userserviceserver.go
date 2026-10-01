package server

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"
)

// UserServiceServer 用户身份与认证服务的 gRPC 适配层。
// 各域方法分散在同包的 user.go / profile.go 等同名文件里。
type UserServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_userv1.UnimplementedUserServiceServer
}

func NewUserServiceServer(svcCtx *svc.ServiceContext) *UserServiceServer {
	return &UserServiceServer{
		svcCtx: svcCtx,
	}
}
