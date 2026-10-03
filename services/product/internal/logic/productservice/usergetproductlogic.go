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

type UserGetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserGetProductLogic {
	return &UserGetProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UserGetProduct 用户端商品详情(仅已上架且未删除)
func (l *UserGetProductLogic) UserGetProduct(in *v1_productv1.UserGetProductReq) (*v1_productv1.UserGetProductResp, error) {
	detail, err := loadUserDetail(l.svcCtx, in.SpuId)
	if err != nil {
		if errors.Is(err, model.ProductNotExist) {
			return &v1_productv1.UserGetProductResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_productv1.UserGetProductResp{Product: converter.ToProtoProductDetailFromUser(detail)}, nil
}
