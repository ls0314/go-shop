package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateCategory 局部更新类目
// 支持改父节点:重新推导 path 并同步改写所有后代的 path
func (l *UpdateCategoryLogic) UpdateCategory(in *v1_productv1.UpdateCategoryReq) (*v1_productv1.UpdateCategoryResp, error) {
	updates, err := decodeUpdateMap(in.UpdatesJson)
	if err != nil {
		return nil, err
	}

	var updated model.SysCategory
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		categoryTxRepo := l.svcCtx.CategoryRepo.WithTx(tx)

		oldCategory, err := categoryTxRepo.GetCategoryById(in.CategoryId)
		if err != nil {
			return err
		}
		if oldCategory == nil {
			return model.CategoryNotExist
		}

		newCategory := *oldCategory
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &newCategory,
		})
		if err != nil {
			return err
		}
		if err := decoder.Decode(updates); err != nil {
			return err
		}

		newPath := oldCategory.CategoryPath
		if newCategory.ParentId != oldCategory.ParentId {
			// 新父类目不能是自己
			if newCategory.ParentId == in.CategoryId {
				return model.CategoryParentInvalid
			}
			if newCategory.ParentId == 0 {
				newPath = fmt.Sprintf("0,%d", in.CategoryId)
			} else {
				parentCategory, err := categoryTxRepo.GetCategoryById(newCategory.ParentId)
				if err != nil {
					return err
				}
				if parentCategory == nil {
					return model.CategoryParentNotExist
				}
				// 新父类目不能是自己的后代(后代 path 以自己的 path 为前缀)
				if strings.HasPrefix(parentCategory.CategoryPath, oldCategory.CategoryPath+",") {
					return model.CategoryParentInvalid
				}
				newPath = fmt.Sprintf("%s,%d", parentCategory.CategoryPath, in.CategoryId)
			}
			newCategory.CategoryPath = newPath
		}

		if err := categoryTxRepo.UpdateCategory(&newCategory); err != nil {
			return err
		}
		// path 变了,所有后代的 path 需同步改写
		if oldCategory.CategoryPath != newPath {
			if err := categoryTxRepo.UpdateChildrenPath(oldCategory.CategoryPath, newPath); err != nil {
				return err
			}
		}

		result, err := categoryTxRepo.GetCategoryById(in.CategoryId)
		if err != nil {
			return err
		}
		if result == nil {
			return model.CategoryNotExist
		}
		updated = *result
		return nil
	})
	if err != nil {
		if isCategoryBizError(err) {
			return &v1_productv1.UpdateCategoryResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	delCategoryCache(l.svcCtx, in.CategoryId)
	// 父节点可能变了,新旧父节点的子树展示都受影响
	delCategoryCache(l.svcCtx, updated.ParentId)

	return &v1_productv1.UpdateCategoryResp{Category: converter.ToProtoCategory(&updated)}, nil
}
