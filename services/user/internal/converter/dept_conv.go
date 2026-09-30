package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoDept 把 model 转成 proto,children 递归转换。
func ToProtoDept(d *model.SysDept) *v1_userv1.Dept {
	if d == nil {
		return nil
	}

	children := make([]*v1_userv1.Dept, 0, len(d.Children))
	for _, c := range d.Children {
		children = append(children, ToProtoDept(c))
	}

	return &v1_userv1.Dept{
		DeptId:    d.DeptId,
		ParentId:  d.ParentId,
		DeptName:  d.DeptName,
		DeptType:  d.DeptType,
		LeaderId:  d.LeaderId,
		SortOrder: d.SortOrder,
		Status:    d.Status,
		CreatedAt: timestamppb.New(d.CreatedAt),
		Children:  children,
	}
}

// ToProtoDeptTree 转换部门树。
func ToProtoDeptTree(list []*model.SysDept) []*v1_userv1.Dept {
	out := make([]*v1_userv1.Dept, 0, len(list))
	for _, d := range list {
		out = append(out, ToProtoDept(d))
	}
	return out
}

// FromProtoDept 把 proto 入参转成 model。created_at 由数据库生成,不从入参取。
func FromProtoDept(p *v1_userv1.Dept) *model.SysDept {
	if p == nil {
		return &model.SysDept{}
	}
	return &model.SysDept{
		DeptId:    p.GetDeptId(),
		ParentId:  p.GetParentId(),
		DeptName:  p.GetDeptName(),
		DeptType:  p.GetDeptType(),
		LeaderId:  p.GetLeaderId(),
		SortOrder: p.GetSortOrder(),
		Status:    p.GetStatus(),
	}
}
