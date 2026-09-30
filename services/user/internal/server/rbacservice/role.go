package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	roleservicelogic "demo-shop/services/user/internal/logic/roleservice"
)

func (s *RBACServiceServer) GetRole(ctx context.Context, in *v1_userv1.GetRoleReq) (*v1_userv1.GetRoleResp, error) {
	l := roleservicelogic.NewGetRoleLogic(ctx, s.svcCtx)
	return l.GetRole(in)
}
func (s *RBACServiceServer) ListRoles(ctx context.Context, in *v1_userv1.ListRolesReq) (*v1_userv1.ListRolesResp, error) {
	l := roleservicelogic.NewListRolesLogic(ctx, s.svcCtx)
	return l.ListRoles(in)
}
func (s *RBACServiceServer) CreateRole(ctx context.Context, in *v1_userv1.CreateRoleReq) (*v1_userv1.CreateRoleResp, error) {
	l := roleservicelogic.NewCreateRoleLogic(ctx, s.svcCtx)
	return l.CreateRole(in)
}
func (s *RBACServiceServer) UpdateRole(ctx context.Context, in *v1_userv1.UpdateRoleReq) (*v1_userv1.UpdateRoleResp, error) {
	l := roleservicelogic.NewUpdateRoleLogic(ctx, s.svcCtx)
	return l.UpdateRole(in)
}
func (s *RBACServiceServer) DeleteRole(ctx context.Context, in *v1_userv1.DeleteRoleReq) (*v1_userv1.DeleteRoleResp, error) {
	l := roleservicelogic.NewDeleteRoleLogic(ctx, s.svcCtx)
	return l.DeleteRole(in)
}
