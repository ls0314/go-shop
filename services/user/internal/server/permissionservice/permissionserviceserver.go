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

var _ v1_userv1.PermissionServiceServer = (*PermissionServiceServer)(nil)

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

func (s *PermissionServiceServer) GetPermission(ctx context.Context, in *v1_userv1.GetPermissionReq) (*v1_userv1.GetPermissionResp, error) {
	l := permissionservicelogic.NewGetPermissionLogic(ctx, s.svcCtx)
	return l.GetPermission(in)
}

func (s *PermissionServiceServer) ListPermissions(ctx context.Context, in *v1_userv1.ListPermissionsReq) (*v1_userv1.ListPermissionsResp, error) {
	l := permissionservicelogic.NewListPermissionsLogic(ctx, s.svcCtx)
	return l.ListPermissions(in)
}

func (s *PermissionServiceServer) CreatePermission(ctx context.Context, in *v1_userv1.CreatePermissionReq) (*v1_userv1.CreatePermissionResp, error) {
	l := permissionservicelogic.NewCreatePermissionLogic(ctx, s.svcCtx)
	return l.CreatePermission(in)
}

func (s *PermissionServiceServer) UpdatePermission(ctx context.Context, in *v1_userv1.UpdatePermissionReq) (*v1_userv1.UpdatePermissionResp, error) {
	l := permissionservicelogic.NewUpdatePermissionLogic(ctx, s.svcCtx)
	return l.UpdatePermission(in)
}

func (s *PermissionServiceServer) DeletePermission(ctx context.Context, in *v1_userv1.DeletePermissionReq) (*v1_userv1.DeletePermissionResp, error) {
	l := permissionservicelogic.NewDeletePermissionLogic(ctx, s.svcCtx)
	return l.DeletePermission(in)
}
