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

type PublishProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishProductLogic {
	return &PublishProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// PublishProduct 上架商品
func (l *PublishProductLogic) PublishProduct(in *v1_productv1.PublishProductReq) (*v1_productv1.PublishProductResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		spuTxRepo := l.svcCtx.ProductRepo.WithTx(tx)

		spu, err := spuTxRepo.GetSpuById(in.SpuId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ProductNotExist
			}
			return err
		}

		// 上架前校验类目:必须是启用状态的叶子类目,否则商品挂在不存在的类目下
		category, err := l.svcCtx.CategoryRepo.GetCategoryById(spu.CategoryId)
		if err != nil {
			return err
		}
		if category == nil || !category.IsLeaf || category.Status != model.CategoryStatusActive {
			return model.ErrCategoryNotUsed
		}

		// 至少一个启用 SKU、有库存、价格 > 0
		if err := spuTxRepo.ValidateSpuPublish(in.SpuId); err != nil {
			return err
		}

		// [FIX-3] 单体此处把 PublishProduct 的错误整个吞掉:
		//   if err := ...; err != nil { if errors.Is(err, gorm.ErrRecordNotFound) { return model.ErrInvalidStatusTransition } }
		// 那个 return 位于内层 if 里、结果被丢弃,函数最终仍返回 nil ——
		// 对已上架商品重复上架会返回"成功"但状态没变。此处如实上报。
		if err := spuTxRepo.PublishProduct(in.SpuId); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ErrInvalidStatusTransition
			}
			return err
		}
		return nil
	})
	if err != nil {
		if isProductBizError(err) {
			return &v1_productv1.PublishProductResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	delProductDetailCache(l.svcCtx, in.SpuId)
	syncProductToES(l.Logger, l.svcCtx, in.SpuId)
	return &v1_productv1.PublishProductResp{}, nil
}
