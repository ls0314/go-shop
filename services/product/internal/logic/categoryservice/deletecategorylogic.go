package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DeleteCategory 删除类目
// 有子类目或已关联商品时拒绝删除;删掉最后一个子节点后父节点恢复为叶子
func (l *DeleteCategoryLogic) DeleteCategory(in *v1_productv1.DeleteCategoryReq) (*v1_productv1.DeleteCategoryResp, error) {
	var parentId int64
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		categoryTxRepo := l.svcCtx.CategoryRepo.WithTx(tx)

		category, err := categoryTxRepo.GetCategoryById(in.CategoryId)
		if err != nil {
			return err
		}
		if category == nil {
			return model.CategoryNotExist
		}
		parentId = category.ParentId

		childCategory, err := categoryTxRepo.GetCategoryByParentId(in.CategoryId)
		if err != nil {
			return err
		}
		if len(childCategory) > 0 {
			return model.CategoryHasChildren
		}

		// 只把"确实查到商品"当业务冲突;查询报错是基础设施故障,不能报成"已关联商品"
		spu, err := l.svcCtx.ProductRepo.GetSpuByCategory(in.CategoryId)
		if err != nil {
			return err
		}
		if spu != nil {
			return model.CategoryHasRel
		}

		if err := categoryTxRepo.DeleteCategoryById(in.CategoryId); err != nil {
			return err
		}

		if category.ParentId != 0 {
			siblings, err := categoryTxRepo.GetCategoryByParentId(category.ParentId)
			if err != nil {
				return err
			}
			if len(siblings) == 0 {
				parent, err := categoryTxRepo.GetCategoryById(category.ParentId)
				if err != nil {
					return err
				}
				// 查不到返回 (nil, nil),判空避免空指针
				if parent == nil {
					return nil
				}
				parent.IsLeaf = true
				return categoryTxRepo.UpdateCategory(parent)
			}
		}

		return nil
	})
	if err != nil {
		if isCategoryBizError(err) {
			return &v1_productv1.DeleteCategoryResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	delCategoryCache(l.svcCtx, in.CategoryId)
	delCategoryCache(l.svcCtx, parentId)
	return &v1_productv1.DeleteCategoryResp{}, nil
}
