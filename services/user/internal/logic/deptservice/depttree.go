package deptservicelogic

import "demo-shop/services/user/internal/model"

// MakeTree 把扁平部门列表组装成树。
// 入参需已按 sort_order 升序,同级顺序即列表顺序。
// 父节点不在入参集合中的节点视为根节点。
func MakeTree(deptList []*model.SysDept) []*model.SysDept {
	deptMap := make(map[int64]*model.SysDept, len(deptList))
	for _, dept := range deptList {
		deptMap[dept.DeptId] = dept
	}

	treeList := make([]*model.SysDept, 0, len(deptList))
	for _, dept := range deptList {
		parent, hasParent := deptMap[dept.ParentId]
		if !hasParent {
			treeList = append(treeList, dept)
			continue
		}
		parent.Children = append(parent.Children, dept)
	}
	return treeList
}
