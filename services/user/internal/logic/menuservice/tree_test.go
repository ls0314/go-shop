package menuservicelogic

import (
	"testing"

	"demo-shop/services/user/internal/model"
)

// menu 构造一个只关心树结构的最小菜单节点
func menu(id, parentId int64) *model.SysMenu {
	return &model.SysMenu{MenuId: id, ParentId: parentId, MenuName: "m"}
}

// collectIds 收集树里所有可达节点的 ID(深度优先,带去重以防万一)
func collectIds(tree []*model.SysMenu) map[int64]bool {
	seen := make(map[int64]bool)
	var walk func(menu *model.SysMenu)
	walk = func(m *model.SysMenu) {
		if seen[m.MenuId] {
			return
		}
		seen[m.MenuId] = true
		for _, child := range m.Children {
			walk(child)
		}
	}
	for _, root := range tree {
		walk(root)
	}
	return seen
}

// TestMakeTreeNormal 正常层级:根节点进 treeList,子节点挂到父节点下
func TestMakeTreeNormal(t *testing.T) {
	tree, warnings := MakeTree([]*model.SysMenu{
		menu(1, 0),
		menu(2, 1),
		menu(3, 1),
		menu(4, 2),
	})

	if len(warnings) != 0 {
		t.Errorf("正常数据不应有告警,实际 = %v", warnings)
	}
	if len(tree) != 1 || tree[0].MenuId != 1 {
		t.Fatalf("根节点应只有 1 个且为 menu_id=1,实际 %d 个", len(tree))
	}
	if len(tree[0].Children) != 2 {
		t.Errorf("menu_id=1 应有 2 个子节点,实际 %d", len(tree[0].Children))
	}
	if got := len(collectIds(tree)); got != 4 {
		t.Errorf("应能到达全部 4 个节点,实际 %d", got)
	}
}

// TestMakeTreeOrphanParent 父节点缺失:该节点本身作为根节点,不应丢失
func TestMakeTreeOrphanParent(t *testing.T) {
	tree, _ := MakeTree([]*model.SysMenu{
		menu(1, 0),
		menu(2, 999), // 999 不在列表里
	})

	if len(tree) != 2 {
		t.Fatalf("父节点缺失的节点应被当作根,期望 2 个根,实际 %d", len(tree))
	}
	if got := len(collectIds(tree)); got != 2 {
		t.Errorf("应能到达全部 2 个节点,实际 %d", got)
	}
}

// TestMakeTreeCycleOnBranch 环挂在正常树上:1→2→3,且 3↔4 互指。
// 环上节点必须仍然可达(修复前的行为是整条环静默消失),并产生告警。
func TestMakeTreeCycleOnBranch(t *testing.T) {
	tree, warnings := MakeTree([]*model.SysMenu{
		menu(1, 0),
		menu(2, 1),
		menu(3, 2),
		menu(4, 3),
		menu(3, 4), // menu_id=3 的父节点是 4,与上一条形成 3↔4
	})

	if len(warnings) == 0 {
		t.Error("成环必须产生告警")
	}
	if got := len(collectIds(tree)); got != 4 {
		t.Errorf("成环时 4 个节点都必须仍可达(修复前会丢),实际可达 %d", got)
	}
}

// TestMakeTreePureCycle 纯环:整条环没有任何节点连到根,DFS 没有入口,
// 必须由兜底逻辑提为根,否则全部菜单消失(前端路由空白)。
func TestMakeTreePureCycle(t *testing.T) {
	tree, warnings := MakeTree([]*model.SysMenu{
		menu(10, 11),
		menu(11, 10), // 10↔11,谁都不是根
	})

	if len(warnings) == 0 {
		t.Error("纯环必须产生告警")
	}
	if got := len(collectIds(tree)); got != 2 {
		t.Errorf("纯环 2 个节点都必须仍可达,实际可达 %d", got)
	}
}

// TestMakeTreeSelfParent 自指:parent_id 指向自己。
// 它既不是别人的子节点也不该消失,按根处理。
func TestMakeTreeSelfParent(t *testing.T) {
	tree, _ := MakeTree([]*model.SysMenu{
		menu(1, 1),
		menu(2, 1),
	})

	if got := len(collectIds(tree)); got != 2 {
		t.Errorf("自指节点与其子节点都必须仍可达,实际可达 %d", got)
	}
}

// TestMakeTreeEmpty 空输入不应 panic
func TestMakeTreeEmpty(t *testing.T) {
	tree, warnings := MakeTree(nil)
	if len(tree) != 0 || len(warnings) != 0 {
		t.Errorf("空输入应返回空树无告警,实际 tree=%d warnings=%d", len(tree), len(warnings))
	}
}
