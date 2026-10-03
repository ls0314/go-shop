package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/utils"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateProduct 局部更新商品(不改 SKU 与图片)。
// 返回合并后的完整详情,调用方无需再查一次。
func (l *UpdateProductLogic) UpdateProduct(in *v1_productv1.UpdateProductReq) (*v1_productv1.UpdateProductResp, error) {
	updates, err := utils.DecodeUpdateJSON(in.UpdatesJson)
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		spuTxRepo := l.svcCtx.ProductRepo.WithTx(tx)

		oldSpu, err := spuTxRepo.GetSpuById(in.SpuId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ProductNotExist
			}
			return err
		}

		// 已上架商品不允许改规格模板
		if updates["spec_template"] != nil && oldSpu.SpuStatus == model.SpuStatusPublished {
			return model.ErrPublishedCantChangeSpec
		}
		// 状态转换只允许 withdrawn → draft
		if newStatus, ok := updates["spu_status"].(string); ok && newStatus != oldSpu.SpuStatus {
			if !(oldSpu.SpuStatus == model.SpuStatusWithdrawn && newStatus == model.SpuStatusDraft) {
				return model.ErrInvalidStatusTransition
			}
		}
		// 目标类目必须存在且为可用的叶子节点。校验直查 DB,不走缓存 ——
		// 缓存里可能是被禁用的旧类目,放过去会写入非法引用。
		if catId, ok := updates["category_id"].(float64); ok {
			cat, err := l.svcCtx.CategoryRepo.GetCategoryById(int64(catId))
			if err != nil {
				return err
			}
			if cat == nil || !cat.IsLeaf || cat.Status != model.CategoryStatusActive {
				return model.ErrCategoryNotUsed
			}
		}

		newSpu := *oldSpu
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &newSpu,
		})
		if err != nil {
			return err
		}
		if err := decoder.Decode(updates); err != nil {
			return err
		}

		return spuTxRepo.UpdateSpu(&newSpu)
	})
	if err != nil {
		if isProductBizError(err) {
			return &v1_productv1.UpdateProductResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	// 事务已提交,才做缓存失效与索引同步
	delProductDetailCache(l.svcCtx, in.SpuId)
	syncProductToES(l.Logger, l.svcCtx, in.SpuId)

	// 复用组装路径(顺带回写缓存),省掉一次"再查一次详情"的 RPC 内往返
	detail, err := loadAdminDetail(l.svcCtx, in.SpuId)
	if err != nil {
		if errors.Is(err, model.ProductNotExist) {
			return &v1_productv1.UpdateProductResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_productv1.UpdateProductResp{Product: converter.ToProtoProductDetail(detail)}, nil
}
