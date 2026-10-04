package server

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	orderservicelogic "demo-shop/services/trade/internal/logic/orderservice"
)

// ============================================================
// OrderService 的 gRPC 适配层
// ============================================================
//
// 本层不含任何业务规则,也不做错误翻译。
//
// 错误约定(与 user / product / marketing 三个服务一致):
//   - 业务失败 → logic 返回的 resp.ErrorMsg 有值,且 err == nil
//   - 基础设施失败 → err != nil,由 gRPC 传成 error
//
// 调用方(trade 的上游是 BFF)按 error 判"要不要重试",
// 按 error_msg 判"给用户看什么"。

// ---- 下单(Saga 编排) ----

func (s *OrderServiceServer) CreateOrder(ctx context.Context, in *v1_tradev1.CreateOrderReq) (*v1_tradev1.CreateOrderResp, error) {
	return orderservicelogic.NewCreateOrderLogic(ctx, s.svcCtx).CreateOrder(in)
}

// ---- 用户端 ----

func (s *OrderServiceServer) ListUserOrders(ctx context.Context, in *v1_tradev1.ListUserOrdersReq) (*v1_tradev1.ListUserOrdersResp, error) {
	return orderservicelogic.NewListUserOrdersLogic(ctx, s.svcCtx).ListUserOrders(in)
}

func (s *OrderServiceServer) GetUserOrder(ctx context.Context, in *v1_tradev1.GetUserOrderReq) (*v1_tradev1.GetUserOrderResp, error) {
	return orderservicelogic.NewGetUserOrderLogic(ctx, s.svcCtx).GetUserOrder(in)
}

func (s *OrderServiceServer) CancelOrder(ctx context.Context, in *v1_tradev1.CancelOrderReq) (*v1_tradev1.CancelOrderResp, error) {
	return orderservicelogic.NewCancelOrderLogic(ctx, s.svcCtx).CancelOrder(in)
}

func (s *OrderServiceServer) ConfirmOrder(ctx context.Context, in *v1_tradev1.ConfirmOrderReq) (*v1_tradev1.ConfirmOrderResp, error) {
	return orderservicelogic.NewConfirmOrderLogic(ctx, s.svcCtx).ConfirmOrder(in)
}

// ---- 管理端 ----

func (s *OrderServiceServer) ListOrders(ctx context.Context, in *v1_tradev1.ListOrdersReq) (*v1_tradev1.ListOrdersResp, error) {
	return orderservicelogic.NewListOrdersLogic(ctx, s.svcCtx).ListOrders(in)
}

func (s *OrderServiceServer) GetOrder(ctx context.Context, in *v1_tradev1.GetOrderReq) (*v1_tradev1.GetOrderResp, error) {
	return orderservicelogic.NewGetOrderLogic(ctx, s.svcCtx).GetOrder(in)
}

func (s *OrderServiceServer) ShipOrder(ctx context.Context, in *v1_tradev1.ShipOrderReq) (*v1_tradev1.ShipOrderResp, error) {
	return orderservicelogic.NewShipOrderLogic(ctx, s.svcCtx).ShipOrder(in)
}
