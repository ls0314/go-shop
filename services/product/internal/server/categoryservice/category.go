package server

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	categoryservicelogic "demo-shop/services/product/internal/logic/categoryservice"
)

func (s *CategoryServiceServer) CreateCategory(ctx context.Context, in *v1_productv1.CreateCategoryReq) (*v1_productv1.CreateCategoryResp, error) {
	l := categoryservicelogic.NewCreateCategoryLogic(ctx, s.svcCtx)
	return l.CreateCategory(in)
}

func (s *CategoryServiceServer) GetCategory(ctx context.Context, in *v1_productv1.GetCategoryReq) (*v1_productv1.GetCategoryResp, error) {
	l := categoryservicelogic.NewGetCategoryLogic(ctx, s.svcCtx)
	return l.GetCategory(in)
}

func (s *CategoryServiceServer) GetCategoryList(ctx context.Context, in *v1_productv1.GetCategoryListReq) (*v1_productv1.GetCategoryListResp, error) {
	l := categoryservicelogic.NewListCategoriesLogic(ctx, s.svcCtx)
	return l.ListCategories(in)
}

func (s *CategoryServiceServer) GetCategoryChildren(ctx context.Context, in *v1_productv1.GetCategoryChildrenReq) (*v1_productv1.GetCategoryChildrenResp, error) {
	l := categoryservicelogic.NewGetCategoryChildrenLogic(ctx, s.svcCtx)
	return l.GetCategoryChildren(in)
}

func (s *CategoryServiceServer) GetCategoryTree(ctx context.Context, in *v1_productv1.GetCategoryTreeReq) (*v1_productv1.GetCategoryTreeResp, error) {
	l := categoryservicelogic.NewGetCategoryTreeLogic(ctx, s.svcCtx)
	return l.GetCategoryTree(in)
}

func (s *CategoryServiceServer) UpdateCategory(ctx context.Context, in *v1_productv1.UpdateCategoryReq) (*v1_productv1.UpdateCategoryResp, error) {
	l := categoryservicelogic.NewUpdateCategoryLogic(ctx, s.svcCtx)
	return l.UpdateCategory(in)
}

func (s *CategoryServiceServer) DeleteCategory(ctx context.Context, in *v1_productv1.DeleteCategoryReq) (*v1_productv1.DeleteCategoryResp, error) {
	l := categoryservicelogic.NewDeleteCategoryLogic(ctx, s.svcCtx)
	return l.DeleteCategory(in)
}
