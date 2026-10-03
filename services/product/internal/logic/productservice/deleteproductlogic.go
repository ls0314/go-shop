package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DeleteProduct 软删除商品并级联软删其 SKU
// TODO:订单关联检查(订单域在 trade 侧,待 C4 拆出后补)
func (l *DeleteProductLogic) DeleteProduct(in *v1_productv1.DeleteProductReq) (*v1_productv1.DeleteProductResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		spuTxRepo := l.svcCtx.ProductRepo.WithTx(tx)

		spu, err := spuTxRepo.GetSpuById(in.SpuId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ProductNotExist
			}
			return err
		}
		if spu.IsDeleted {
			return model.ProductNotExist
		}

		if err := spuTxRepo.BatchDeleteSkuById(in.SpuId); err != nil {
			return err
		}
		return spuTxRepo.DeleteSpuById(in.SpuId)
	})
	if err != nil {
		if isProductBizError(err) {
			return &v1_productv1.DeleteProductResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	// [FIX-8] 单体在事务返回后无条件清缓存与 ES(不看 err),事务回滚时会把
	// 一个仍然存在的商品从索引里抹掉。此处仅在事务确实提交后才执行旁路清理。
	delProductDetailCache(l.svcCtx, in.SpuId)
	removeProductFromES(l.Logger, l.svcCtx, in.SpuId)
	return &v1_productv1.DeleteProductResp{}, nil
}
