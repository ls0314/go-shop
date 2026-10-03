package inventoryservice

import (
	"context"
	"demo-shop/api/gen/product/v1"
	inventoryservicelogic "demo-shop/services/product/internal/logic/inventoryservice"
)

func (s *InventoryServiceServer) LockStock(ctx context.Context, in *v1_productv1.LockStockReq) (*v1_productv1.LockStockResp, error) {
	return inventoryservicelogic.NewLockStockLogic(ctx, s.svcCtx).LockStock(in)
}

func (s *InventoryServiceServer) DeductStock(ctx context.Context, in *v1_productv1.DeductStockReq) (*v1_productv1.DeductStockResp, error) {
	return inventoryservicelogic.NewDeductStockLogic(ctx, s.svcCtx).DeductStock(in)
}

func (s *InventoryServiceServer) ReleaseStock(ctx context.Context, in *v1_productv1.ReleaseStockReq) (*v1_productv1.ReleaseStockResp, error) {
	return inventoryservicelogic.NewReleaseStockLogic(ctx, s.svcCtx).ReleaseStock(in)
}

func (s *InventoryServiceServer) RefundStock(ctx context.Context, in *v1_productv1.RefundStockReq) (*v1_productv1.RefundStockResp, error) {
	return inventoryservicelogic.NewRefundStockLogic(ctx, s.svcCtx).RefundStock(in)
}

// ---- 查询(管理端库存看板) ----

func (s *InventoryServiceServer) GetSkuStock(ctx context.Context, in *v1_productv1.GetSkuStockReq) (*v1_productv1.GetSkuStockResp, error) {
	return inventoryservicelogic.NewGetSkuStockLogic(ctx, s.svcCtx).GetSkuStock(in)
}

func (s *InventoryServiceServer) GetSkuStockList(ctx context.Context, in *v1_productv1.GetSkuStockListReq) (*v1_productv1.GetSkuStockListResp, error) {
	return inventoryservicelogic.NewGetSkuStockListLogic(ctx, s.svcCtx).GetSkuStockList(in)
}

func (s *InventoryServiceServer) GetWarnStockList(ctx context.Context, in *v1_productv1.GetWarnStockListReq) (*v1_productv1.GetWarnStockListResp, error) {
	return inventoryservicelogic.NewGetWarnStockListLogic(ctx, s.svcCtx).GetWarnStockList(in)
}

// ---- 流水查询(管理端库存流水页) ----

func (s *InventoryServiceServer) ListStockLogs(ctx context.Context, in *v1_productv1.ListStockLogsReq) (*v1_productv1.ListStockLogsResp, error) {
	return inventoryservicelogic.NewListStockLogsLogic(ctx, s.svcCtx).ListStockLogs(in)
}

// ---- 手动调整(管理端) ----

func (s *InventoryServiceServer) AdjustStock(ctx context.Context, in *v1_productv1.AdjustStockReq) (*v1_productv1.AdjustStockResp, error) {
	return inventoryservicelogic.NewAdjustStockLogic(ctx, s.svcCtx).AdjustStock(in)
}
