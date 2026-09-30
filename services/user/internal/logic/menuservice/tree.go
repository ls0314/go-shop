package menuservicelogic

import "demo-shop/services/user/internal/model"

// MakeTree 把扁平菜单列表组装成树。
func MakeTree(menuList []*model.SysMenu) []*model.SysMenu {
	menuMap := make(map[int64]*model.SysMenu, len(menuList))
	for _, menu := range menuList {
		menuMap[menu.MenuId] = menu
	}

	treeList := make([]*model.SysMenu, 0, len(menuList))
	for _, menu := range menuList {
		parent, hasParent := menuMap[menu.ParentId]
		if !hasParent {
			treeList = append(treeList, menu)
			continue
		}
		parent.Children = append(parent.Children, menu)
	}
	return treeList
}
