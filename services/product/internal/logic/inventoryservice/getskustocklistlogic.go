package inventoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetSkuStockListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSkuStockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkuStockListLogic {
	return &GetSkuStockListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetSkuStockList 一个 SPU 下全部未删除 SKU 的库存 + 聚合值。
// 聚合值(total_stock/total_lock/total_sold)由本服务算:
// 单体时代是遍历 SKU 累加,这段逻辑属于库存域,不该留在 HTTP 层。
func (l *GetSkuStockListLogic) GetSkuStockList(in *v1_productv1.GetSkuStockListReq) (*v1_productv1.GetSkuStockListResp, error) {
	skuList, err := l.svcCtx.ProductRepo.GetSkuListBySpuId(in.SpuId)
	if err != nil {
		return nil, err
	}

	// SPU 不存在时,单体时代 GetSpuById 会返回 ErrRecordNotFound → 整个请求失败。
	// 保留该行为:否则前端会拿到一个"所有 SKU 都没有 spu_name"的响应,更难排查。
	spu, err := l.svcCtx.ProductRepo.GetSpuById(in.SpuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_productv1.GetSkuStockListResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}

	items := make([]*v1_productv1.SkuStock, 0, len(skuList))
	var totalStock, totalLock, totalSold int64
	for i := range skuList {
		sku := &skuList[i]
		totalStock += sku.Stock
		totalLock += sku.LockStock
		totalSold += sku.SoldCount
		items = append(items, converter.ToProtoSkuStock(sku, spu.SpuName))
	}

	return &v1_productv1.GetSkuStockListResp{
		Items:      items,
		TotalStock: totalStock,
		TotalLock:  totalLock,
		TotalSold:  totalSold,
	}, nil
}
