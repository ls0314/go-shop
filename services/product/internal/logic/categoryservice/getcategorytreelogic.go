package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryTreeLogic {
	return &GetCategoryTreeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetCategoryTree 获取类目树。
// makeCategoryTree 会检测父节点成环,返回的告警在此记录 —— 环上的类目
// 会被提为根节点而不丢失,但数据本身需要人工修。
func (l *GetCategoryTreeLogic) GetCategoryTree(in *v1_productv1.GetCategoryTreeReq) (*v1_productv1.GetCategoryTreeResp, error) {
	categoryList, err := l.svcCtx.CategoryRepo.GetAllCategory(in.Level, in.IncludeDisabled)
	if err != nil {
		return nil, err
	}

	tree, warnings := makeCategoryTree(categoryList)
	if len(warnings) > 0 {
		l.Errorf("类目树数据异常: %s", formatCategoryTreeWarnings(warnings))
	}

	return &v1_productv1.GetCategoryTreeResp{
		Items: converter.ToProtoCategoryTree(tree),
	}, nil
}
