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

type GetSkuStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSkuStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkuStockLogic {
	return &GetSkuStockLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetSkuStock 单个 SKU 的库存详情(管理端库存看板)。
// 返回 DB 口径而非闸门实时值:管理端要的是"账面库存 + 锁定数 + 销量",
// 闸门键只存可用量,拿不到另外两个数。
func (l *GetSkuStockLogic) GetSkuStock(in *v1_productv1.GetSkuStockReq) (*v1_productv1.GetSkuStockResp, error) {
	sku, err := l.svcCtx.ProductRepo.GetSku(in.SkuId)
	if err != nil {
		if errors.Is(err, model.ErrSkuNotExist) {
			return &v1_productv1.GetSkuStockResp{ErrorMsg: model.ErrSkuNotExist.Error()}, nil
		}
		return nil, err
	}

	spu, err := l.svcCtx.ProductRepo.GetSpuById(sku.SpuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_productv1.GetSkuStockResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}

	return &v1_productv1.GetSkuStockResp{
		Sku: converter.ToProtoSkuStock(sku, spu.SpuName),
	}, nil
}
