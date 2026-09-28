package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	permissionservicelogic "demo-shop/services/user/internal/logic/permissionservice"
	"demo-shop/services/user/internal/svc"
)

type PermissionServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_userv1.UnimplementedPermissionServiceServer
}

func NewPermissionServiceServer(svcCtx *svc.ServiceContext) *PermissionServiceServer {
	return &PermissionServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *PermissionServiceServer) GetPermVersion(ctx context.Context, in *v1_userv1.GetPermVersionReq) (*v1_userv1.GetPermVersionResp, error) {
	l := permissionservicelogic.NewGetPermVersionLogic(ctx, s.svcCtx)
	return l.GetPermVersion(in)
}

func (s *PermissionServiceServer) ListPermCodesByUserId(ctx context.Context, in *v1_userv1.ListPermCodesByUserIdReq) (*v1_userv1.ListPermCodesByUserIdResp, error) {
	l := permissionservicelogic.NewListPermCodesByUserIdLogic(ctx, s.svcCtx)
	return l.ListPermCodesByUserId(in)
}

func (s *PermissionServiceServer) ListPermCodesByApi(ctx context.Context, in *v1_userv1.ListPermCodesByApiReq) (*v1_userv1.ListPermCodesByApiResp, error) {
	l := permissionservicelogic.NewListPermCodesByApiLogic(ctx, s.svcCtx)
	return l.ListPermCodesByApi(in)
}
