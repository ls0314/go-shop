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

type UpdateProductFullLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductFullLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFullLogic {
	return &UpdateProductFullLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateProductFull 全量更新商品:SPU 基本信息 + SKU 增改删 + 图片增改删
func (l *UpdateProductFullLogic) UpdateProductFull(in *v1_productv1.UpdateProductFullReq) (*v1_productv1.UpdateProductFullResp, error) {
	req := converter.ToFullUpdateProductReq(in)

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		spuTxRepo := l.svcCtx.ProductRepo.WithTx(tx)

		oldSpu, err := spuTxRepo.GetSpuById(in.SpuId)
		if err != nil {
			// [FIX-13] 单体此处返回 gorm.ErrRecordNotFound 原始错误,
			// 而 UpdateProduct/DeleteProduct 返回 ProductNotExist,同一语义两种错误。
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ProductNotExist
			}
			return err
		}

		if oldSpu.SpuStatus == model.SpuStatusPublished && req.SpecTemplate != nil {
			return model.ErrPublishedCantChangeSpec
		}

		newSpu := *oldSpu
		if req.SpuName != "" {
			newSpu.SpuName = req.SpuName
		}
		if req.CategoryId != nil {
			newSpu.CategoryId = *req.CategoryId
		}
		if req.Brand != "" {
			newSpu.Brand = req.Brand
		}
		if req.Description != "" {
			newSpu.Description = req.Description
		}
		if req.MainImage != "" {
			newSpu.MainImage = req.MainImage
		}
		if req.SpecTemplate != nil {
			newSpu.SpecTemplate = *req.SpecTemplate
		}
		if req.Priority != nil {
			newSpu.Priority = *req.Priority
		}
		if err := spuTxRepo.UpdateSpu(&newSpu); err != nil {
			return err
		}

		// SKU:有 id 更新、无 id 新增、库有而请求无则软删
		if req.SkuList != nil {
			if len(*req.SkuList) == 0 {
				return model.ErrSkuListEmpty
			}
			// checkSpecValue=false:全量更新允许保留原有规格组合,
			// 查库判重会把"本次未改动的同一个 SKU"误判成冲突
			if err := validateSpecSku(in.SpuId, newSpu.SpecTemplate, *req.SkuList, spuTxRepo, false); err != nil {
				return err
			}

			keepIds := make([]int64, 0, len(*req.SkuList))
			for _, sku := range *req.SkuList {
				if sku.SkuId > 0 {
					keepIds = append(keepIds, sku.SkuId)
				}
			}
			if err := spuTxRepo.SoftDeleteSkusExcept(in.SpuId, keepIds); err != nil {
				return err
			}

			for _, sku := range *req.SkuList {
				sku := sku
				sku.SpuId = in.SpuId
				if sku.SkuId > 0 {
					if err := spuTxRepo.UpdateSku(&sku); err != nil {
						return err
					}
				} else {
					if err := spuTxRepo.CreateSku(&sku); err != nil {
						return err
					}
				}
			}
		}

		// 图片:先删指定 id,再按有无 image_id 增改
		if err := spuTxRepo.DeleteImageByIds(req.DeleteImageIds); err != nil {
			return err
		}
		if req.ImageList != nil {
			for _, img := range *req.ImageList {
				img := img
				img.SpuId = in.SpuId
				if img.ImageId > 0 {
					if err := spuTxRepo.UpdateImage(&img); err != nil {
						return err
					}
				} else {
					if err := spuTxRepo.CreateImage(&img); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		if isProductBizError(err) {
			return &v1_productv1.UpdateProductFullResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	delProductDetailCache(l.svcCtx, in.SpuId)
	syncProductToES(l.Logger, l.svcCtx, in.SpuId)

	detail, err := loadAdminDetail(l.svcCtx, in.SpuId)
	if err != nil {
		if errors.Is(err, model.ProductNotExist) {
			return &v1_productv1.UpdateProductFullResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_productv1.UpdateProductFullResp{Product: converter.ToProtoProductDetail(detail)}, nil
}
