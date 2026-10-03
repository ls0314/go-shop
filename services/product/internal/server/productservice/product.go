package server

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	productservicelogic "demo-shop/services/product/internal/logic/productservice"
)

// ---- 管理端 ----

func (s *ProductServiceServer) CreateProduct(ctx context.Context, in *v1_productv1.CreateProductReq) (*v1_productv1.CreateProductResp, error) {
	return productservicelogic.NewCreateProductLogic(ctx, s.svcCtx).CreateProduct(in)
}

func (s *ProductServiceServer) GetProductList(ctx context.Context, in *v1_productv1.GetProductListReq) (*v1_productv1.GetProductListResp, error) {
	return productservicelogic.NewGetProductListLogic(ctx, s.svcCtx).GetProductList(in)
}

func (s *ProductServiceServer) GetProduct(ctx context.Context, in *v1_productv1.GetProductReq) (*v1_productv1.GetProductResp, error) {
	return productservicelogic.NewGetProductLogic(ctx, s.svcCtx).GetProduct(in)
}

func (s *ProductServiceServer) UpdateProduct(ctx context.Context, in *v1_productv1.UpdateProductReq) (*v1_productv1.UpdateProductResp, error) {
	return productservicelogic.NewUpdateProductLogic(ctx, s.svcCtx).UpdateProduct(in)
}

func (s *ProductServiceServer) UpdateProductFull(ctx context.Context, in *v1_productv1.UpdateProductFullReq) (*v1_productv1.UpdateProductFullResp, error) {
	return productservicelogic.NewUpdateProductFullLogic(ctx, s.svcCtx).UpdateProductFull(in)
}

func (s *ProductServiceServer) PublishProduct(ctx context.Context, in *v1_productv1.PublishProductReq) (*v1_productv1.PublishProductResp, error) {
	return productservicelogic.NewPublishProductLogic(ctx, s.svcCtx).PublishProduct(in)
}

func (s *ProductServiceServer) WithdrawProduct(ctx context.Context, in *v1_productv1.WithdrawProductReq) (*v1_productv1.WithdrawProductResp, error) {
	return productservicelogic.NewWithdrawProductLogic(ctx, s.svcCtx).WithdrawProduct(in)
}

func (s *ProductServiceServer) DeleteProduct(ctx context.Context, in *v1_productv1.DeleteProductReq) (*v1_productv1.DeleteProductResp, error) {
	return productservicelogic.NewDeleteProductLogic(ctx, s.svcCtx).DeleteProduct(in)
}

// ---- 用户端 ----

func (s *ProductServiceServer) UserGetProductList(ctx context.Context, in *v1_productv1.UserGetProductListReq) (*v1_productv1.UserGetProductListResp, error) {
	return productservicelogic.NewUserGetProductListLogic(ctx, s.svcCtx).UserGetProductList(in)
}

func (s *ProductServiceServer) UserGetProduct(ctx context.Context, in *v1_productv1.UserGetProductReq) (*v1_productv1.UserGetProductResp, error) {
	return productservicelogic.NewUserGetProductLogic(ctx, s.svcCtx).UserGetProduct(in)
}

// ---- SKU / SPU 级读接口 ----

func (s *ProductServiceServer) GetSku(ctx context.Context, in *v1_productv1.GetSkuReq) (*v1_productv1.GetSkuResp, error) {
	return productservicelogic.NewGetSkuLogic(ctx, s.svcCtx).GetSku(in)
}

func (s *ProductServiceServer) BatchGetSkus(ctx context.Context, in *v1_productv1.BatchGetSkusReq) (*v1_productv1.BatchGetSkusResp, error) {
	return productservicelogic.NewBatchGetSkusLogic(ctx, s.svcCtx).BatchGetSkus(in)
}

func (s *ProductServiceServer) GetSpu(ctx context.Context, in *v1_productv1.GetSpuReq) (*v1_productv1.GetSpuResp, error) {
	return productservicelogic.NewGetSpuLogic(ctx, s.svcCtx).GetSpu(in)
}
