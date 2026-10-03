package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetCategory 查询类目详情。查不到返回 CategoryNotExist
func (l *GetCategoryLogic) GetCategory(in *v1_productv1.GetCategoryReq) (*v1_productv1.GetCategoryResp, error) {
	// 注意:不能让 repo 的错误无条件变成"类目不存在" —— 数据库故障会被误报成业务失败
	category, err := l.svcCtx.CategoryRepo.GetCategoryById(in.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &v1_productv1.GetCategoryResp{ErrorMsg: model.CategoryNotExist.Error()}, nil
	}
	return &v1_productv1.GetCategoryResp{Category: converter.ToProtoCategory(category)}, nil
}
