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

type GetSpuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSpuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSpuLogic {
	return &GetSpuLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetSpu 按 SPU ID 查 SPU 快照(含类目名)。
// 与 GetProduct 的区别:不带 SKU/图片列表,用于只要商品基本信息的校验路径。
func (l *GetSpuLogic) GetSpu(in *v1_productv1.GetSpuReq) (*v1_productv1.GetSpuResp, error) {
	spu, err := l.svcCtx.ProductRepo.GetSpuById(in.SpuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_productv1.GetSpuResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}

	resp := &v1_productv1.GetSpuResp{
		Spu: &v1_productv1.Spu{
			SpuId:        spu.SpuId,
			SpuName:      spu.SpuName,
			CategoryId:   spu.CategoryId,
			Brand:        spu.Brand,
			Description:  spu.Description,
			MainImage:    spu.MainImage,
			SpecTemplate: string(spu.SpecTemplate),
			SpuStatus:    spu.SpuStatus,
			Priority:     spu.Priority,
			IsDeleted:    spu.IsDeleted,
		},
	}

	// 类目名是展示字段:查不到不报错,留空即可(不像列表页那样依赖它排序/筛选)
	category, err := getCategory(l.svcCtx, spu.CategoryId)
	if err != nil {
		return nil, err
	}
	if category != nil {
		resp.CategoryName = category.CategoryName
	}

	return resp, nil
}
