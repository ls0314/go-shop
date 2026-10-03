package categoryservicelogic

import (
	"demo-shop/services/product/internal/model"
	"strings"
	"testing"
)

// 按 (category_id, parent_id) 造扁平类目列表,模拟 sys_category 的查询结果
func cats(pairs ...[2]int64) []*model.SysCategory {
	out := make([]*model.SysCategory, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, &model.SysCategory{
			CategoryId:   p[0],
			ParentId:     p[1],
			CategoryName: "cat",
			IsVisible:    true,
			Status:       "active",
		})
	}
	return out
}

// collectIds 前序遍历收集树上的所有节点 id。
// 用 seen 去重:环上的节点被提为根后仍可能挂在别的节点下,
// 直接递归会顺着环无限展开(测试里表现为栈溢出)。
func collectIds(nodes []*model.GetTreeCategoryResp) []int64 {
	var out []int64
	seen := map[int64]bool{}
	var walk func([]*model.GetTreeCategoryResp)
	walk = func(ns []*model.GetTreeCategoryResp) {
		for _, n := range ns {
			if seen[n.CategoryId] {
				continue
			}
			seen[n.CategoryId] = true
			out = append(out, n.CategoryId)
			walk(n.Children)
		}
	}
	walk(nodes)
	return out
}

func hasWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

// 正常三层树:1 → 2 → 3,外加根节点 4
func TestMakeCategoryTree_Normal(t *testing.T) {
	tree, warnings := makeCategoryTree(cats(
		[2]int64{1, 0},
		[2]int64{2, 1},
		[2]int64{3, 2},
		[2]int64{4, 0},
	))

	if len(warnings) != 0 {
		t.Fatalf("正常数据不应产生告警, got %v", warnings)
	}
	if len(tree) != 2 {
		t.Fatalf("根节点应为 2 个, got %d", len(tree))
	}
	if got := collectIds(tree); len(got) != 4 {
		t.Fatalf("树上应含全部 4 个节点, got %v", got)
	}
	// 1 下面挂 2,2 下面挂 3
	if len(tree[0].Children) != 1 || tree[0].Children[0].CategoryId != 2 {
		t.Fatalf("节点 1 的子节点应为 [2], got %+v", tree[0].Children)
	}
	if len(tree[0].Children[0].Children) != 1 ||
		tree[0].Children[0].Children[0].CategoryId != 3 {
		t.Fatalf("节点 2 的子节点应为 [3], got %+v", tree[0].Children[0].Children)
	}
}

// 父节点悬空(被删/被 level 过滤断链):按根节点处理,节点不丢
func TestMakeCategoryTree_DanglingParent(t *testing.T) {
	tree, warnings := makeCategoryTree(cats(
		[2]int64{1, 0},
		[2]int64{5, 99}, // 99 不在结果集中
	))

	if len(warnings) != 0 {
		t.Fatalf("悬空父节点按根处理,不产生告警, got %v", warnings)
	}
	if len(tree) != 2 {
		t.Fatalf("悬空节点应被提为根,根数应为 2, got %d", len(tree))
	}
	if got := collectIds(tree); len(got) != 2 {
		t.Fatalf("节点不应丢失, got %v", got)
	}
}

// 自指(id == parent_id):必须提为根,否则永远没有入口
func TestMakeCategoryTree_SelfReference(t *testing.T) {
	tree, warnings := makeCategoryTree(cats(
		[2]int64{1, 0},
		[2]int64{7, 7},
	))

	if len(warnings) != 0 {
		t.Fatalf("自指按根处理即可,无需告警, got %v", warnings)
	}
	if got := collectIds(tree); len(got) != 2 {
		t.Fatalf("自指节点不应丢失, got %v", got)
	}
}

// 环挂在正常子树上:环上被回溯到的那一环提为根,节点不丢且有告警
func TestMakeCategoryTree_CycleAttachedToTree(t *testing.T) {
	// 1 → 2 → 3 → 2(环:2 与 3 互为父子)
	tree, warnings := makeCategoryTree(cats(
		[2]int64{1, 0},
		[2]int64{2, 1},
		[2]int64{3, 2},
		[2]int64{2, 3},
	))

	if !hasWarning(warnings, "存在环") {
		t.Fatalf("环上数据必须告警, got %v", warnings)
	}
	if got := collectIds(tree); len(got) != 3 {
		t.Fatalf("环上节点不应丢失(应提为根), got %v", got)
	}
}

// 纯环:整条环没有任何节点连到根,DFS 没有入口,靠兜底提根
func TestMakeCategoryTree_PureCycle(t *testing.T) {
	tree, warnings := makeCategoryTree(cats(
		[2]int64{1, 0},
		[2]int64{8, 9},
		[2]int64{9, 8},
	))

	if !hasWarning(warnings, "不在任何根节点的子树上") {
		t.Fatalf("纯环必须走兜底告警, got %v", warnings)
	}
	if got := collectIds(tree); len(got) != 3 {
		t.Fatalf("纯环节点不应丢失, got %v", got)
	}
}

// 空输入不应 panic
func TestMakeCategoryTree_Empty(t *testing.T) {
	tree, warnings := makeCategoryTree(nil)
	if len(tree) != 0 || len(warnings) != 0 {
		t.Fatalf("空输入应返回空树与空告警, got %v / %v", tree, warnings)
	}
}
