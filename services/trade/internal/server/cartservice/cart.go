package server

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	cartservicelogic "demo-shop/services/trade/internal/logic/cartservice"
)

// ============================================================
// CartService 的 gRPC 适配层(纯转发)
// ============================================================
//
// 错误约定与 OrderService 一致:
//   - 业务失败 → resp.ErrorMsg 有值,err == nil
//   - 基础设施失败 → err != nil

func (s *CartServiceServer) CreateCartItem(ctx context.Context, in *v1_tradev1.CreateCartItemReq) (*v1_tradev1.CreateCartItemResp, error) {
	return cartservicelogic.NewCreateCartItemLogic(ctx, s.svcCtx).CreateCartItem(in)
}

func (s *CartServiceServer) GetCartItem(ctx context.Context, in *v1_tradev1.GetCartItemReq) (*v1_tradev1.GetCartItemResp, error) {
	return cartservicelogic.NewGetCartItemLogic(ctx, s.svcCtx).GetCartItem(in)
}

func (s *CartServiceServer) ListCartItems(ctx context.Context, in *v1_tradev1.ListCartItemsReq) (*v1_tradev1.ListCartItemsResp, error) {
	return cartservicelogic.NewListCartItemsLogic(ctx, s.svcCtx).ListCartItems(in)
}

func (s *CartServiceServer) UpdateCartItem(ctx context.Context, in *v1_tradev1.UpdateCartItemReq) (*v1_tradev1.UpdateCartItemResp, error) {
	return cartservicelogic.NewUpdateCartItemLogic(ctx, s.svcCtx).UpdateCartItem(in)
}

func (s *CartServiceServer) DeleteCartItem(ctx context.Context, in *v1_tradev1.DeleteCartItemReq) (*v1_tradev1.DeleteCartItemResp, error) {
	return cartservicelogic.NewDeleteCartItemLogic(ctx, s.svcCtx).DeleteCartItem(in)
}

func (s *CartServiceServer) SelectAllCartItems(ctx context.Context, in *v1_tradev1.SelectAllCartItemsReq) (*v1_tradev1.SelectAllCartItemsResp, error) {
	return cartservicelogic.NewSelectAllCartItemsLogic(ctx, s.svcCtx).SelectAllCartItems(in)
}

func (s *CartServiceServer) GetCartItemCount(ctx context.Context, in *v1_tradev1.GetCartItemCountReq) (*v1_tradev1.GetCartItemCountResp, error) {
	return cartservicelogic.NewGetCartItemCountLogic(ctx, s.svcCtx).GetCartItemCount(in)
}

func (s *CartServiceServer) GetCartPayPreview(ctx context.Context, in *v1_tradev1.GetCartPayPreviewReq) (*v1_tradev1.GetCartPayPreviewResp, error) {
	return cartservicelogic.NewGetCartPayPreviewLogic(ctx, s.svcCtx).GetCartPayPreview(in)
}
