package inventoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// 低库存预警的默认值,与单体 inventory_service.go 一致
const defaultWarnThreshold = 10

type GetWarnStockListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWarnStockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWarnStockListLogic {
	return &GetWarnStockListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetWarnStockList 低于阈值的可用库存列表(管理端预警),按库存升序。
// 默认值与单体一致:threshold<=0 → 10,spu_status 空 → published。
func (l *GetWarnStockListLogic) GetWarnStockList(in *v1_productv1.GetWarnStockListReq) (*v1_productv1.GetWarnStockListResp, error) {
	threshold := int(in.Threshold)
	if threshold <= 0 {
		threshold = defaultWarnThreshold
	}
	spuStatus := in.SpuStatus
	if spuStatus == "" {
		spuStatus = model.SpuStatusPublished
	}

	rows, err := l.svcCtx.ProductRepo.GetWarnStock(threshold, spuStatus)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_productv1.SkuStock, 0, len(rows))
	for i := range rows {
		items = append(items, converter.ToProtoSkuStockFromWarnItem(&rows[i]))
	}

	return &v1_productv1.GetWarnStockListResp{Items: items}, nil
}
