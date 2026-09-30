package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	menuservicelogic "demo-shop/services/user/internal/logic/menuservice"
)

func (s *RBACServiceServer) GetMenu(ctx context.Context, in *v1_userv1.GetMenuReq) (*v1_userv1.GetMenuResp, error) {
	l := menuservicelogic.NewGetMenuLogic(ctx, s.svcCtx)
	return l.GetMenu(in)
}
func (s *RBACServiceServer) ListMenus(ctx context.Context, in *v1_userv1.ListMenusReq) (*v1_userv1.ListMenusResp, error) {
	l := menuservicelogic.NewListMenusLogic(ctx, s.svcCtx)
	return l.ListMenus(in)
}
func (s *RBACServiceServer) CreateMenu(ctx context.Context, in *v1_userv1.CreateMenuReq) (*v1_userv1.CreateMenuResp, error) {
	l := menuservicelogic.NewCreateMenuLogic(ctx, s.svcCtx)
	return l.CreateMenu(in)
}
func (s *RBACServiceServer) UpdateMenu(ctx context.Context, in *v1_userv1.UpdateMenuReq) (*v1_userv1.UpdateMenuResp, error) {
	l := menuservicelogic.NewUpdateMenuLogic(ctx, s.svcCtx)
	return l.UpdateMenu(in)
}
func (s *RBACServiceServer) DeleteMenu(ctx context.Context, in *v1_userv1.DeleteMenuReq) (*v1_userv1.DeleteMenuResp, error) {
	l := menuservicelogic.NewDeleteMenuLogic(ctx, s.svcCtx)
	return l.DeleteMenu(in)
}
func (s *RBACServiceServer) GetMenuTreeByUserId(ctx context.Context, in *v1_userv1.GetMenuTreeByUserIdReq) (*v1_userv1.GetMenuTreeByUserIdResp, error) {
	l := menuservicelogic.NewGetMenuTreeByUserIdLogic(ctx, s.svcCtx)
	return l.GetMenuTreeByUserId(in)
}
func (s *RBACServiceServer) GetMenuTreeByRoleId(ctx context.Context, in *v1_userv1.GetMenuTreeByRoleIdReq) (*v1_userv1.GetMenuTreeByRoleIdResp, error) {
	l := menuservicelogic.NewGetMenuTreeByRoleIdLogic(ctx, s.svcCtx)
	return l.GetMenuTreeByRoleId(in)
}
