package categoryservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/utils"
	"errors"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 类目域可预期的业务失败。落在此集合内以 error_msg 返回(gRPC error 为 nil),
// 其余视为基础设施故障,gRPC error 非 nil 由调用方重试。
var categoryBizErrors = []error{
	model.CategoryDisable,
	model.CategoryUkExist,
	model.CategoryParentNotExist,
	model.CategoryLevelDeep,
	model.CategoryNotExist,
	model.CategoryHasChildren,
	model.CategoryHasRel,
	model.CategoryParentInvalid,
}

func isCategoryBizError(err error) bool {
	for _, target := range categoryBizErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// toListCategoryResp 类目实体 → 列表项
func toListCategoryResp(c *model.SysCategory) model.GetListCategoryResp {
	return model.GetListCategoryResp{
		CategoryId:    c.CategoryId,
		ParentId:      c.ParentId,
		CategoryName:  c.CategoryName,
		CategoryLevel: c.CategoryLevel,
		SortOrder:     c.SortOrder,
		IsLeaf:        c.IsLeaf,
		IsVisible:     c.IsVisible,
		Status:        c.Status,
	}
}

// decodeUpdateMap 把 proto 传过来的 JSON 对象文本解成更新字段映射。
// 实现在 utils.DecodeUpdateJSON,与商品域共用(契约语义必须一致)。
func decodeUpdateMap(updatesJson string) (map[string]interface{}, error) {
	return utils.DecodeUpdateJSON(updatesJson)
}

// formatCategoryTreeWarnings 把告警拼成一行,便于日志输出
func formatCategoryTreeWarnings(warnings []string) string {
	return utils.JoinWarnings(warnings)
}

// delCategoryCache 失效单个类目详情缓存。
// Redis 是共享实例,写路径必须清缓存,否则最长 1 小时内读到旧类目。
func delCategoryCache(svcCtx *svc.ServiceContext, categoryId int64) {
	if svcCtx.Redis == nil || categoryId == 0 {
		return
	}
	_, _ = svcCtx.Redis.Del(fmt.Sprintf("category:%d", categoryId))
}

// makeCategoryTree 把扁平类目列表组装成树,并返回数据异常告警。
//
// 正常情况下 ParentId 指向的类目都在同一列表内;出现以下两种说明 sys_category
// 数据有问题,需人工修:
//   - 悬空:ParentId 指向的类目不存在 —— 按根节点处理;
//   - 成环:沿着 ParentId 上溯能回到自己 —— 环上不存在真正的根节点。
//     此时把环上被回溯到的那一环提为根,让它和子树仍然可见,
//     而不是整条环都进不了结果(类目从前端树上静默消失)。
func makeCategoryTree(categoryList []*model.SysCategory) ([]*model.GetTreeCategoryResp, []string) {
	categoryMap := make(map[int64]*model.GetTreeCategoryResp, len(categoryList))
	nodes := make([]*model.GetTreeCategoryResp, 0, len(categoryList))
	for _, c := range categoryList {
		node := &model.GetTreeCategoryResp{
			CategoryId:    c.CategoryId,
			CategoryName:  c.CategoryName,
			CategoryLevel: c.CategoryLevel,
			IsVisible:     c.IsVisible,
			Status:        c.Status,
			ParentId:      c.ParentId,
			Children:      []*model.GetTreeCategoryResp{},
		}
		categoryMap[c.CategoryId] = node
		nodes = append(nodes, node)
	}

	const (
		visiting = 1
		done     = 2
	)
	state := make(map[int64]int, len(nodes))
	var warnings []string

	// 建父子关系;父节点缺失(含按 level 过滤造成的断链)与自指都当根
	tree := make([]*model.GetTreeCategoryResp, 0, len(nodes))
	for _, node := range nodes {
		parent, hasParent := categoryMap[node.ParentId]
		if !hasParent || parent == node {
			tree = append(tree, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	// 从根 DFS;visiting 表示还在当前递归路径上,再遇到即成环
	var walk func(node *model.GetTreeCategoryResp)
	walk = func(node *model.GetTreeCategoryResp) {
		if state[node.CategoryId] == done {
			return
		}
		state[node.CategoryId] = visiting
		kept := make([]*model.GetTreeCategoryResp, 0, len(node.Children))
		for _, child := range node.Children {
			switch state[child.CategoryId] {
			case visiting:
				warnings = append(warnings, fmt.Sprintf(
					"类目树存在环:category_id=%d(parent_id=%d) 的父节点链回到自身,已提为根节点",
					child.CategoryId, child.ParentId))
				tree = append(tree, child)
				// 不放进 kept:不再挂在当前节点下,避免同一节点两个入口
			case done:
				kept = append(kept, child)
			default:
				kept = append(kept, child)
				walk(child)
			}
		}
		node.Children = kept
		state[node.CategoryId] = done
	}
	// 索引循环:提为根的节点会追加到 tree 末尾,必须一并走到
	for i := 0; i < len(tree); i++ {
		walk(tree[i])
	}

	// 兜底:整条都是环、环上没有任何节点连到根时,DFS 没有入口
	for _, node := range nodes {
		if state[node.CategoryId] == done {
			continue
		}
		tree = append(tree, node)
		warnings = append(warnings, fmt.Sprintf(
			"类目树存在环:category_id=%d(parent_id=%d) 不在任何根节点的子树上,已提为根节点",
			node.CategoryId, node.ParentId))
		walk(node)
	}

	return tree, warnings
}

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateCategory 创建类目
// 同一父节点下类目名唯一;level 与 path 由父节点推导
func (l *CreateCategoryLogic) CreateCategory(in *v1_productv1.CreateCategoryReq) (*v1_productv1.CreateCategoryResp, error) {
	category := converter.FromProtoCategory(in.Category)

	existing, err := l.svcCtx.CategoryRepo.GetCategoryUk(category.ParentId, category.CategoryName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &v1_productv1.CreateCategoryResp{ErrorMsg: model.CategoryUkExist.Error()}, nil
	}

	// 父类目:根节点用内存构造值,非根节点必须查得到
	parentCategory := &model.SysCategory{}
	if category.ParentId == 0 {
		parentCategory.CategoryLevel = 0
		parentCategory.CategoryPath = "0"
		parentCategory.Status = model.CategoryStatusActive
		parentCategory.IsLeaf = false
	} else {
		parentCategory, err = l.svcCtx.CategoryRepo.GetCategoryById(category.ParentId)
		if err != nil {
			return nil, err
		}
		// GetCategoryById 查不到返回 (nil, nil),必须判 nil,否则下面读字段会 panic
		if parentCategory == nil {
			return &v1_productv1.CreateCategoryResp{ErrorMsg: model.CategoryParentNotExist.Error()}, nil
		}
	}

	if parentCategory.Status == "disabled" {
		return &v1_productv1.CreateCategoryResp{ErrorMsg: model.CategoryDisable.Error()}, nil
	}
	if parentCategory.CategoryLevel >= 4 {
		return &v1_productv1.CreateCategoryResp{ErrorMsg: model.CategoryLevelDeep.Error()}, nil
	}

	category.CategoryLevel = parentCategory.CategoryLevel + 1
	category.IsLeaf = true
	if category.Status == "" {
		category.Status = model.CategoryStatusActive
	}
	// path 要拿到自增主键后才能算,先占位
	category.CategoryPath = "0"

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		categoryTxRepo := l.svcCtx.CategoryRepo.WithTx(tx)

		if err := categoryTxRepo.CreateCategory(category); err != nil {
			return err
		}
		category.CategoryPath = parentCategory.CategoryPath + "," + strconv.FormatInt(category.CategoryId, 10)
		if err := categoryTxRepo.UpdateCategory(category); err != nil {
			return err
		}
		// 父类目原本是叶子,新增子节点后不再
		if category.ParentId != 0 && parentCategory.IsLeaf {
			parentCategory.IsLeaf = false
			if err := categoryTxRepo.UpdateCategory(parentCategory); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 父类目的 is_leaf 变了,其缓存一并失效
	delCategoryCache(l.svcCtx, category.ParentId)

	return &v1_productv1.CreateCategoryResp{
		CategoryId:   category.CategoryId,
		CategoryPath: category.CategoryPath,
	}, nil
}
