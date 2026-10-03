package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateProduct 创建商品:同一事务内创建 SPU、SKU 列表与图片列表。
// 事务提交后才做 ES 同步等旁路动作 —— 回滚时不得把商品提前写进索引。
func (l *CreateProductLogic) CreateProduct(in *v1_productv1.CreateProductReq) (*v1_productv1.CreateProductResp, error) {
	spu := converter.FromProtoSpu(in.Spu)

	category, err := l.svcCtx.CategoryRepo.GetCategoryById(spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil || category.Status != model.CategoryStatusActive || !category.IsLeaf {
		return &v1_productv1.CreateProductResp{ErrorMsg: model.ErrCategoryNotUsed.Error()}, nil
	}

	if spu.SkuList == nil || len(*spu.SkuList) == 0 {
		return &v1_productv1.CreateProductResp{ErrorMsg: model.ErrSkuListEmpty.Error()}, nil
	}
	// checkSpecValue=true:创建路径要查库确认规格组合未被同 SPU 的其他 SKU 占用
	if err := validateSpecSku(0, spu.SpecTemplate, *spu.SkuList, l.svcCtx.ProductRepo, true); err != nil {
		return l.createErrResp(err)
	}

	var spuId int64
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := l.svcCtx.ProductRepo.WithTx(tx)

		newSpuId, err := txRepo.CreateSpu(spu)
		if err != nil {
			return err
		}
		spuId = newSpuId

		for _, sku := range *spu.SkuList {
			sku := sku
			sku.SpuId = spuId
			if sku.Price <= 0 {
				return model.ErrInvalidPrice
			}
			// [FIX-7] 单体写的是 err != nil || skuCode != nil,把"查询报错"也判成
			// SKU 编码重复,数据库一抖用户就看到"SKU编码需唯一"。此处只把
			// "确实查到同码 SKU"当冲突,真故障原样上抛。
			existingSku, err := txRepo.GetSkuBySkuCode(sku.SkuCode)
			if err != nil {
				return err
			}
			if existingSku != nil {
				return model.ErrSkuCodeNotOnly
			}
			if err := txRepo.CreateSku(&sku); err != nil {
				return err
			}
		}

		if spu.ImageList != nil {
			for _, image := range *spu.ImageList {
				image := image
				image.SpuId = spuId
				if err := txRepo.CreateImage(&image); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return l.createErrResp(err)
	}

	syncProductToES(l.Logger, l.svcCtx, spuId)
	return &v1_productv1.CreateProductResp{SpuId: spuId}, nil
}

// createErrResp 业务失败 → error_msg;基础设施故障 → gRPC error。
// 失败时一律回 spu_id=0:事务内已回填过自增主键,不回零会让调用方看到一个并不存在的 ID。
func (l *CreateProductLogic) createErrResp(err error) (*v1_productv1.CreateProductResp, error) {
	if isProductBizError(err) {
		return &v1_productv1.CreateProductResp{ErrorMsg: err.Error()}, nil
	}
	return nil, err
}
