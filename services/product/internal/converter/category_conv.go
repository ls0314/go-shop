package converter

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/model"
)

// ToProtoCategory 把 model 转成 proto。
func ToProtoCategory(c *model.SysCategory) *v1_productv1.Category {
	if c == nil {
		return nil
	}
	return &v1_productv1.Category{
		CategoryId:    c.CategoryId,
		ParentId:      c.ParentId,
		CategoryName:  c.CategoryName,
		CategoryLevel: c.CategoryLevel,
		CategoryPath:  c.CategoryPath,
		SortOrder:     c.SortOrder,
		IconUrl:       c.IconUrl,
		IsLeaf:        c.IsLeaf,
		IsVisible:     c.IsVisible,
		Status:        c.Status,
	}
}

// FromProtoCategory 把 proto 入参转成 model。
// created_at/updated_at 由数据库生成,不从入参取。
func FromProtoCategory(p *v1_productv1.Category) *model.SysCategory {
	if p == nil {
		return &model.SysCategory{}
	}
	return &model.SysCategory{
		CategoryId:    p.GetCategoryId(),
		ParentId:      p.GetParentId(),
		CategoryName:  p.GetCategoryName(),
		CategoryLevel: p.GetCategoryLevel(),
		CategoryPath:  p.GetCategoryPath(),
		SortOrder:     p.GetSortOrder(),
		IconUrl:       p.GetIconUrl(),
		IsLeaf:        p.GetIsLeaf(),
		IsVisible:     p.GetIsVisible(),
		Status:        p.GetStatus(),
	}
}

// ToProtoCategoryList 转换类目列表项。
func ToProtoCategoryList(items []model.GetListCategoryResp) []*v1_productv1.CategoryListItem {
	out := make([]*v1_productv1.CategoryListItem, 0, len(items))
	for _, it := range items {
		out = append(out, &v1_productv1.CategoryListItem{
			CategoryId:    it.CategoryId,
			ParentId:      it.ParentId,
			CategoryName:  it.CategoryName,
			CategoryLevel: it.CategoryLevel,
			SortOrder:     it.SortOrder,
			IsLeaf:        it.IsLeaf,
			IsVisible:     it.IsVisible,
			Status:        it.Status,
		})
	}
	return out
}

// ToProtoCategoryTree 转换类目树,children 递归转换。
func ToProtoCategoryTree(list []*model.GetTreeCategoryResp) []*v1_productv1.CategoryTreeNode {
	out := make([]*v1_productv1.CategoryTreeNode, 0, len(list))
	for _, n := range list {
		out = append(out, &v1_productv1.CategoryTreeNode{
			CategoryId:    n.CategoryId,
			CategoryName:  n.CategoryName,
			CategoryLevel: n.CategoryLevel,
			IsVisible:     n.IsVisible,
			Status:        n.Status,
			Children:      ToProtoCategoryTree(n.Children),
		})
	}
	return out
}
