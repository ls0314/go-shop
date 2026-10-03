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

type WithdrawProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWithdrawProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawProductLogic {
	return &WithdrawProductLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// WithdrawProduct 下架商品(仅 published 可下架)
func (l *WithdrawProductLogic) WithdrawProduct(in *v1_productv1.WithdrawProductReq) (*v1_productv1.WithdrawProductResp, error) {
	// 先确认商品存在,便于把"不存在"与"状态不允许下架"区分开
	_, err := l.svcCtx.ProductRepo.GetSpuById(in.SpuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_productv1.WithdrawProductResp{ErrorMsg: model.ProductNotExist.Error()}, nil
		}
		return nil, err
	}

	if err := l.svcCtx.ProductRepo.WithdrawProduct(in.SpuId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_productv1.WithdrawProductResp{ErrorMsg: model.ErrInvalidStatusTransition.Error()}, nil
		}
		return nil, err
	}

	delProductDetailCache(l.svcCtx, in.SpuId)
	// 下架后商品不该再出现在搜索结果里,直接删文档(而非索引成 withdrawn)
	removeProductFromES(l.Logger, l.svcCtx, in.SpuId)
	return &v1_productv1.WithdrawProductResp{}, nil
}
