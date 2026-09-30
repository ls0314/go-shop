package server

import (
	"context"
	"demo-shop/api/gen/user/v1"
	deptservicelogic "demo-shop/services/user/internal/logic/deptservice"
)

func (s *RBACServiceServer) GetDept(ctx context.Context, in *v1_userv1.GetDeptReq) (*v1_userv1.GetDeptResp, error) {
	l := deptservicelogic.NewGetDeptLogic(ctx, s.svcCtx)
	return l.GetDept(in)
}
func (s *RBACServiceServer) ListDepts(ctx context.Context, in *v1_userv1.ListDeptsReq) (*v1_userv1.ListDeptsResp, error) {
	l := deptservicelogic.NewListDeptsLogic(ctx, s.svcCtx)
	return l.ListDepts(in)
}
func (s *RBACServiceServer) CreateDept(ctx context.Context, in *v1_userv1.CreateDeptReq) (*v1_userv1.CreateDeptResp, error) {
	l := deptservicelogic.NewCreateDeptLogic(ctx, s.svcCtx)
	return l.CreateDept(in)
}
func (s *RBACServiceServer) UpdateDept(ctx context.Context, in *v1_userv1.UpdateDeptReq) (*v1_userv1.UpdateDeptResp, error) {
	l := deptservicelogic.NewUpdateDeptLogic(ctx, s.svcCtx)
	return l.UpdateDept(in)
}
func (s *RBACServiceServer) DeleteDept(ctx context.Context, in *v1_userv1.DeleteDeptReq) (*v1_userv1.DeleteDeptResp, error) {
	l := deptservicelogic.NewDeleteDeptLogic(ctx, s.svcCtx)
	return l.DeleteDept(in)
}
func (s *RBACServiceServer) GetDeptTreeByUserId(ctx context.Context, in *v1_userv1.GetDeptTreeByUserIdReq) (*v1_userv1.GetDeptTreeByUserIdResp, error) {
	l := deptservicelogic.NewGetDeptTreeByUserIdLogic(ctx, s.svcCtx)
	return l.GetDeptTreeByUserId(in)
}
