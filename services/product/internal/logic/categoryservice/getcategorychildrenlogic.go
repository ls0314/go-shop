package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryChildrenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryChildrenLogic {
	return &GetCategoryChildrenLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetCategoryChildren 查询类目下的直接子类目。父类目不存在时报 CategoryNotExist
func (l *GetCategoryChildrenLogic) GetCategoryChildren(in *v1_productv1.GetCategoryChildrenReq) (*v1_productv1.GetCategoryChildrenResp, error) {
	parent, err := l.svcCtx.CategoryRepo.GetCategoryById(in.ParentId)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return &v1_productv1.GetCategoryChildrenResp{ErrorMsg: model.CategoryNotExist.Error()}, nil
	}

	children, err := l.svcCtx.CategoryRepo.GetCategoryByParentId(in.ParentId)
	if err != nil {
		return nil, err
	}

	items := make([]model.GetListCategoryResp, 0, len(children))
	for i := range children {
		items = append(items, toListCategoryResp(&children[i]))
	}
	return &v1_productv1.GetCategoryChildrenResp{
		Items: converter.ToProtoCategoryList(items),
	}, nil
}
