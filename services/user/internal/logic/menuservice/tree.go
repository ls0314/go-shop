package menuservicelogic

import (
	"fmt"
	"strings"

	"demo-shop/services/user/internal/model"
)

// MakeTree 把扁平菜单列表组装成树,并返回数据异常告警。
//
// 正常情况下 ParentId 指向的菜单一定在同一列表内(整棵树一起查出来),
// 且整棵树有根(root 的 ParentId 不在列表里),所以从根出发就能走到所有节点。
// 出现以下两种情况说明 sys_menu 的数据有问题,需要人工修:
//   - 悬空:ParentId 指向的菜单不存在 —— 该节点自己当根;
//   - 成环:沿着 ParentId 上溯能回到自己 —— 环上没有根,从根出发走不到,
//     整条环会从前端动态路由里静默消失。这里把环上节点提为根,保证仍然可见。
func MakeTree(menuList []*model.SysMenu) ([]*model.SysMenu, []string) {
	menuMap := make(map[int64]*model.SysMenu, len(menuList))
	for _, menu := range menuList {
		menuMap[menu.MenuId] = menu
	}

	const (
		visiting = 1
		done     = 2
	)
	state := make(map[int64]int, len(menuList))
	var warnings []string

	// 建父子关系。自指(id == parent_id)必须当根,否则它既是自己的子节点、
	// 又没有任何入口,整棵子树都走不到。
	treeList := make([]*model.SysMenu, 0, len(menuList))
	for _, menu := range menuList {
		parent, hasParent := menuMap[menu.ParentId]
		if !hasParent || parent == menu {
			treeList = append(treeList, menu)
			continue
		}
		parent.Children = append(parent.Children, menu)
	}

	// 从根出发做 DFS。visiting = 还在当前递归路径上,再遇到就是环。
	// 发现环时立刻把它从父节点子列表摘掉并提为根 —— 必须当场摘并当场入 treeList,
	// 这样它自己还能带着子树被走到。
	var walk func(menu *model.SysMenu)
	walk = func(menu *model.SysMenu) {
		if state[menu.MenuId] == done {
			return
		}
		state[menu.MenuId] = visiting
		kept := make([]*model.SysMenu, 0, len(menu.Children))
		for _, child := range menu.Children {
			switch state[child.MenuId] {
			case visiting:
				warnings = append(warnings, fmt.Sprintf(
					"菜单树存在环:menu_id=%d(parent_id=%d) 的父节点链回到自身,已提为根节点",
					child.MenuId, child.ParentId))
				treeList = append(treeList, child)
				// 不放进 kept —— 不再挂在当前节点下,避免同一节点有两个入口
			case done:
				kept = append(kept, child)
			default:
				kept = append(kept, child)
				walk(child)
			}
		}
		// 原地替换(不能用 children[:0] 复用底层数组,那会覆盖同层其他节点)
		menu.Children = kept
		state[menu.MenuId] = done
	}
	// 索引循环:提为根的节点会追加到 treeList 末尾,必须一并走到
	for i := 0; i < len(treeList); i++ {
		walk(treeList[i])
	}

	// 兜底:整条都是环、环上没有任何节点连到根时,DFS 压根没有入口。
	// 把没走到的节点提为根并递归检查,保证没有任何菜单被静默丢弃。
	for _, menu := range menuList {
		if state[menu.MenuId] == done {
			continue
		}
		treeList = append(treeList, menu)
		warnings = append(warnings, fmt.Sprintf(
			"菜单树存在环:menu_id=%d(parent_id=%d) 不在任何根节点的子树上,已提为根节点",
			menu.MenuId, menu.ParentId))
		walk(menu)
	}

	return treeList, warnings
}

// FormatTreeWarnings 把告警拼成一行,便于日志输出
func FormatTreeWarnings(warnings []string) string {
	return strings.Join(warnings, "; ")
}
