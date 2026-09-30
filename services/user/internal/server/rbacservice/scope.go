package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	scopeservicelogic "demo-shop/services/user/internal/logic/scopeservice"
)

func (s *RBACServiceServer) GetScope(ctx context.Context, in *v1_userv1.GetScopeReq) (*v1_userv1.GetScopeResp, error) {
	l := scopeservicelogic.NewGetScopeLogic(ctx, s.svcCtx)
	return l.GetScope(in)
}
func (s *RBACServiceServer) ListScopes(ctx context.Context, in *v1_userv1.ListScopesReq) (*v1_userv1.ListScopesResp, error) {
	l := scopeservicelogic.NewListScopesLogic(ctx, s.svcCtx)
	return l.ListScopes(in)
}
func (s *RBACServiceServer) CreateScope(ctx context.Context, in *v1_userv1.CreateScopeReq) (*v1_userv1.CreateScopeResp, error) {
	l := scopeservicelogic.NewCreateScopeLogic(ctx, s.svcCtx)
	return l.CreateScope(in)
}
func (s *RBACServiceServer) UpdateScope(ctx context.Context, in *v1_userv1.UpdateScopeReq) (*v1_userv1.UpdateScopeResp, error) {
	l := scopeservicelogic.NewUpdateScopeLogic(ctx, s.svcCtx)
	return l.UpdateScope(in)
}
func (s *RBACServiceServer) DeleteScope(ctx context.Context, in *v1_userv1.DeleteScopeReq) (*v1_userv1.DeleteScopeResp, error) {
	l := scopeservicelogic.NewDeleteScopeLogic(ctx, s.svcCtx)
	return l.DeleteScope(in)
}
