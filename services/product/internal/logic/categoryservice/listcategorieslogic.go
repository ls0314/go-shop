package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListCategories 分页查询类目
func (l *ListCategoriesLogic) ListCategories(in *v1_productv1.GetCategoryListReq) (*v1_productv1.GetCategoryListResp, error) {
	page, pageSize := int(in.Page), int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	// <=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	categoryList, total, err := l.svcCtx.CategoryRepo.GetCategoryList(page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.GetListCategoryResp, 0, len(categoryList))
	for _, c := range categoryList {
		items = append(items, toListCategoryResp(c))
	}

	return &v1_productv1.GetCategoryListResp{
		Items: converter.ToProtoCategoryList(items),
		Total: total,
	}, nil
}
