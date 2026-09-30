package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
)

// ToProtoMenu 把 model 转成 proto,children 递归转换。
func ToProtoMenu(m *model.SysMenu) *v1_userv1.Menu {
	if m == nil {
		return nil
	}

	children := make([]*v1_userv1.Menu, 0, len(m.Children))
	for _, c := range m.Children {
		children = append(children, ToProtoMenu(c))
	}

	return &v1_userv1.Menu{
		MenuId:    m.MenuId,
		ParentId:  m.ParentId,
		MenuName:  m.MenuName,
		MenuType:  m.MenuType,
		Icon:      m.Icon,
		RoutePath: m.RoutePath,
		Component: m.ComponentPath,
		IsVisible: m.IsVisible,
		IsCache:   m.IsCache,
		SortOrder: m.SortOrder,
		MetaInfo:  toStruct(m.MetaInfo),
		CreatedAt: timestamppb.New(m.CreatedAt),
		Children:  children,
	}
}

// ToProtoMenuTree 转换菜单树。
func ToProtoMenuTree(list []*model.SysMenu) []*v1_userv1.Menu {
	out := make([]*v1_userv1.Menu, 0, len(list))
	for _, m := range list {
		out = append(out, ToProtoMenu(m))
	}
	return out
}

// FromProtoMenu 把 proto 入参转成 model。
// meta_info 为只读字段,不从入参取;created_at 由数据库生成。
func FromProtoMenu(p *v1_userv1.Menu) *model.SysMenu {
	if p == nil {
		return &model.SysMenu{}
	}
	var meta datatypes.JSONMap
	if p.MetaInfo != nil {
		meta = datatypes.JSONMap(p.MetaInfo.AsMap())
	}
	return &model.SysMenu{
		MenuId:        p.GetMenuId(),
		ParentId:      p.GetParentId(),
		MenuName:      p.GetMenuName(),
		MenuType:      p.GetMenuType(),
		Icon:          p.GetIcon(),
		RoutePath:     p.GetRoutePath(),
		ComponentPath: p.GetComponent(),
		IsVisible:     p.GetIsVisible(),
		IsCache:       p.GetIsCache(),
		SortOrder:     p.GetSortOrder(),
		MetaInfo:      meta,
	}
}

// toStruct 把 jsonb 列转成 proto Struct。类型不支持时返回 nil 而非报错,
// meta_info 是展示用字段,缺失不应让整个查询失败。
func toStruct(m datatypes.JSONMap) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}
