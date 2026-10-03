package productservicelogic

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

type GetSkuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkuLogic {
	return &GetSkuLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetSku 按 SKU ID 查 SKU 与归属 SPU 的快照 + 可用性判定位。
// 供购物车加购校验、结算校验、库存查询使用。
func (l *GetSkuLogic) GetSku(in *v1_productv1.GetSkuReq) (*v1_productv1.GetSkuResp, error) {
	sku, err := l.svcCtx.ProductRepo.GetSku(in.SkuId)
	if err != nil {
		// repo 已把 gorm.ErrRecordNotFound 翻成 ErrSkuNotExist,文案是跨服务契约
		if errors.Is(err, model.ErrSkuNotExist) {
			return &v1_productv1.GetSkuResp{ErrorMsg: model.ErrSkuNotExist.Error()}, nil
		}
		return nil, err
	}

	spu, err := l.svcCtx.ProductRepo.GetSpuById(sku.SpuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// SKU 在、SPU 被硬删:数据不一致,如实报"商品不存在"而非 500
			return &v1_productv1.GetSkuResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}

	return converter.ToProtoGetSkuResp(sku, spu), nil
}
