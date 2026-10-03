package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetProduct 管理端商品详情(缓存优先,命中后仍用闸门键校正库存)
func (l *GetProductLogic) GetProduct(in *v1_productv1.GetProductReq) (*v1_productv1.GetProductResp, error) {
	detail, err := loadAdminDetail(l.svcCtx, in.SpuId)
	if err != nil {
		if errors.Is(err, model.ProductNotExist) {
			return &v1_productv1.GetProductResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_productv1.GetProductResp{Product: converter.ToProtoProductDetail(detail)}, nil
}
