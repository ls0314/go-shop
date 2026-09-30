package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	menupermservicelogic "demo-shop/services/user/internal/logic/menupermservice"
	rolemenuservicelogic "demo-shop/services/user/internal/logic/rolemenuservice"
	rolepermservicelogic "demo-shop/services/user/internal/logic/rolepermservice"
	userdeptservicelogic "demo-shop/services/user/internal/logic/userdeptservice"
	userroleservicelogic "demo-shop/services/user/internal/logic/userroleservice"
)

// ---- 角色-权限绑定 ----

func (s *RBACServiceServer) AssignRolePerms(ctx context.Context, in *v1_userv1.AssignRolePermsReq) (*v1_userv1.AssignRolePermsResp, error) {
	l := rolepermservicelogic.NewAssignRolePermsLogic(ctx, s.svcCtx)
	return l.AssignRolePerms(in)
}
func (s *RBACServiceServer) ListRolePerms(ctx context.Context, in *v1_userv1.ListRolePermsReq) (*v1_userv1.ListRolePermsResp, error) {
	l := rolepermservicelogic.NewListRolePermsLogic(ctx, s.svcCtx)
	return l.ListRolePerms(in)
}
func (s *RBACServiceServer) ClearRolePerms(ctx context.Context, in *v1_userv1.ClearRolePermsReq) (*v1_userv1.ClearRolePermsResp, error) {
	l := rolepermservicelogic.NewClearRolePermsLogic(ctx, s.svcCtx)
	return l.ClearRolePerms(in)
}

// ---- 角色-菜单绑定 ----

func (s *RBACServiceServer) AssignRoleMenus(ctx context.Context, in *v1_userv1.AssignRoleMenusReq) (*v1_userv1.AssignRoleMenusResp, error) {
	l := rolemenuservicelogic.NewAssignRoleMenusLogic(ctx, s.svcCtx)
	return l.AssignRoleMenus(in)
}
func (s *RBACServiceServer) ListRoleMenus(ctx context.Context, in *v1_userv1.ListRoleMenusReq) (*v1_userv1.ListRoleMenusResp, error) {
	l := rolemenuservicelogic.NewListRoleMenusLogic(ctx, s.svcCtx)
	return l.ListRoleMenus(in)
}
func (s *RBACServiceServer) ClearRoleMenus(ctx context.Context, in *v1_userv1.ClearRoleMenusReq) (*v1_userv1.ClearRoleMenusResp, error) {
	l := rolemenuservicelogic.NewClearRoleMenusLogic(ctx, s.svcCtx)
	return l.ClearRoleMenus(in)
}

// ---- 菜单-权限绑定 ----

func (s *RBACServiceServer) AssignMenuPerms(ctx context.Context, in *v1_userv1.AssignMenuPermsReq) (*v1_userv1.AssignMenuPermsResp, error) {
	l := menupermservicelogic.NewAssignMenuPermsLogic(ctx, s.svcCtx)
	return l.AssignMenuPerms(in)
}
func (s *RBACServiceServer) ListMenuPerms(ctx context.Context, in *v1_userv1.ListMenuPermsReq) (*v1_userv1.ListMenuPermsResp, error) {
	l := menupermservicelogic.NewListMenuPermsLogic(ctx, s.svcCtx)
	return l.ListMenuPerms(in)
}
func (s *RBACServiceServer) ClearMenuPerms(ctx context.Context, in *v1_userv1.ClearMenuPermsReq) (*v1_userv1.ClearMenuPermsResp, error) {
	l := menupermservicelogic.NewClearMenuPermsLogic(ctx, s.svcCtx)
	return l.ClearMenuPerms(in)
}

// ---- 用户-角色绑定 ----

func (s *RBACServiceServer) AssignUserRoles(ctx context.Context, in *v1_userv1.AssignUserRolesReq) (*v1_userv1.AssignUserRolesResp, error) {
	l := userroleservicelogic.NewAssignUserRolesLogic(ctx, s.svcCtx)
	return l.AssignUserRoles(in)
}
func (s *RBACServiceServer) ListUserRoles(ctx context.Context, in *v1_userv1.ListUserRolesReq) (*v1_userv1.ListUserRolesResp, error) {
	l := userroleservicelogic.NewListUserRolesLogic(ctx, s.svcCtx)
	return l.ListUserRoles(in)
}
func (s *RBACServiceServer) ClearUserRoles(ctx context.Context, in *v1_userv1.ClearUserRolesReq) (*v1_userv1.ClearUserRolesResp, error) {
	l := userroleservicelogic.NewClearUserRolesLogic(ctx, s.svcCtx)
	return l.ClearUserRoles(in)
}

// ---- 用户-部门绑定 ----

func (s *RBACServiceServer) AssignUserDepts(ctx context.Context, in *v1_userv1.AssignUserDeptsReq) (*v1_userv1.AssignUserDeptsResp, error) {
	l := userdeptservicelogic.NewAssignUserDeptsLogic(ctx, s.svcCtx)
	return l.AssignUserDepts(in)
}
func (s *RBACServiceServer) ListUserDepts(ctx context.Context, in *v1_userv1.ListUserDeptsReq) (*v1_userv1.ListUserDeptsResp, error) {
	l := userdeptservicelogic.NewListUserDeptsLogic(ctx, s.svcCtx)
	return l.ListUserDepts(in)
}
func (s *RBACServiceServer) ClearUserDepts(ctx context.Context, in *v1_userv1.ClearUserDeptsReq) (*v1_userv1.ClearUserDeptsResp, error) {
	l := userdeptservicelogic.NewClearUserDeptsLogic(ctx, s.svcCtx)
	return l.ClearUserDepts(in)
}
