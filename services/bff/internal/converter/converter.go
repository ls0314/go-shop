// Package converter proto 消息 → .api 生成的 types.*。
//
// ============================================================
// 为什么现在才建这个包
// ============================================================
//
// 写 user 域时这些映射只有 1 个包用,放在 logic/user/convert.go 里
// 就够了(那个包内私有)。
//
// 但 role / permission / menu / scope / dept 五个包需要的是**同一批**
// 映射(Role → RoleItem、Permission → PermissionItem、Menu → MenuItem、
// Dept → DeptItem)。Go 的包内标识符不跨包,故只有两条路:
//
//	A. 五个包各抄一份   → 同一份映射逻辑存在 5 处,
//	                      改一处忘一处必然发生,且症状是
//	                      "某个域的接口少返回一个字段",极难发现
//	B. 抽成共享包       → 一份实现,六个域共用  ← 采用
//
// 判据是**复用面**:1 个包用 → 放 logic 里;≥2 个包用 → 抽出来。
// 不是"提前抽象",而是复用面确实到了。
//
// ============================================================
// 本包只放**跨域共用**的映射
// ============================================================
//
// 只被单个域用到的(如 FieldUpdate 构造、分页兜底)留在各自的 logic 包里 ——
// 放进来会让这个包变成一个"什么都往里丢"的杂物间。
//
// ============================================================
// 为什么每个实体有两个函数(单个 / 切片)
// ============================================================
//
// toRoleItem 用于 Get/Create/Update(返回单个);
// toRoleItems 用于 List(返回切片)。
//
// 切片版本**必须用 make(..., 0, len)** 而不是 var x []T:
// 后者在无数据时是 nil,JSON 序列化成 null,而前端有
// `list.length` / `items.map(...)` 这类写法,null 会直接抛错。
// proto 的 repeated 字段在无数据时返回空切片,但显式兜一次
// 能防住"服务端返回 nil 切片"的边界。
package converter

import (
	"time"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/types"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================================
// 时间字段
// ============================================================
//
// proto 的时间字段是 google.protobuf.Timestamp,而 .api 里的 types.*
// 把时间声明成 string。
//
// 这不是随意的:单体那边的响应是 gin 直接序列化 proto struct,
// 而 timestamppb.Timestamp 实现了 MarshalJSON → 输出 **RFC3339 字符串**:
//
//	"2026-02-14T10:30:00Z"
//
// 前端解的就是这个形状。若 .api 里声明成 int64(时间戳),
// 前端展示时间的地方会全部变成一串数字 —— 那是破坏契约。
//
// 故这里显式做同样的格式化,保证 BFF 与单体的输出逐字一致。

// Timestamp timestamppb.Timestamp → RFC3339 字符串。
//
// **nil 安全**:proto 里未设置的时间字段是 nil 指针,直接调
// AsTime() 会 panic。返回空串与单体一致。
func Timestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	// UTC + RFC3339,与 timestamppb.MarshalJSON 的输出格式一致。
	//
	// 用 time.RFC3339 而不是 RFC3339Nano:后者在纳秒为 0 时会省掉
	// 小数部分,而 timestamppb 永远输出纳秒精度 —— 差异会让
	// "同一时间在两个接口里长得不一样"。
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// ============================================================
// Role
// ============================================================

func RoleItem(r *v1_userv1.Role) types.RoleItem {
	return types.RoleItem{
		RoleId:      r.GetRoleId(),
		RoleName:    r.GetRoleName(),
		RoleType:    r.GetRoleType(),
		Description: r.GetDescription(),
		IsSystem:    r.GetIsSystem(),
		IsDefault:   r.GetIsDefault(),
		DataScope:   r.GetDataScope(),
		CreatedAt:   Timestamp(r.GetCreatedAt()),
	}
}

func RoleItems(rs []*v1_userv1.Role) []types.RoleItem {
	out := make([]types.RoleItem, 0, len(rs))
	for _, r := range rs {
		out = append(out, RoleItem(r))
	}
	return out
}

// ============================================================
// Permission
// ============================================================

func PermissionItem(p *v1_userv1.Permission) types.PermissionItem {
	return types.PermissionItem{
		PermissionId:   p.GetPermissionId(),
		PermissionCode: p.GetPermissionCode(),
		PermissionName: p.GetPermissionName(),
		PermissionType: p.GetPermissionType(),
		RequestMethod:  p.GetRequestMethod(),
		ApiPath:        p.GetApiPath(),
		Description:    p.GetDescription(),
		IsSystem:       p.GetIsSystem(),
	}
}

func PermissionItems(ps []*v1_userv1.Permission) []types.PermissionItem {
	out := make([]types.PermissionItem, 0, len(ps))
	for _, p := range ps {
		out = append(out, PermissionItem(p))
	}
	return out
}

// ============================================================
// Menu(递归)
// ============================================================
//
// Children 是递归的:GetMenuTreeByUserId / GetMenuTreeByRoleId 返回的是
// **已组好树的**结构(proto 里 children 已填充),故映射必须递归 ——
// 只映射顶层会让前端渲染动态路由时只能看到一级菜单。

func MenuItem(m *v1_userv1.Menu) types.MenuItem {
	return types.MenuItem{
		MenuId:    m.GetMenuId(),
		ParentId:  m.GetParentId(),
		MenuName:  m.GetMenuName(),
		MenuType:  m.GetMenuType(),
		Icon:      m.GetIcon(),
		RoutePath: m.GetRoutePath(),
		Component: m.GetComponent(),
		IsVisible: m.GetIsVisible(),
		IsCache:   m.GetIsCache(),
		SortOrder: m.GetSortOrder(),
		// MetaInfo 是 google.protobuf.Struct,而 types.MenuItem.MetaInfo
		// 是 interface{}(.api 里为什么用 interface{} 而不是 any:
		// goctl 1.9.2 的语法解析器只认 interface{})。
		//
		// **必须走 AsMap()**:直接塞 structpb.Struct 会让 encoding/json
		// 输出 {"fields": {...}} 这种 protobuf 内部结构,而不是前端要的
		// 普通对象 —— 前端解的是 meta_info.title / meta_info.icon
		// 这类直接取值,拿到 fields 包装会全变 undefined。
		//
		// AsMap() 对 nil 接收者安全(返回 nil map),序列化成 null,
		// 与 proto 未设置时的语义一致。
		MetaInfo:  m.GetMetaInfo().AsMap(),
		CreatedAt: Timestamp(m.GetCreatedAt()),
		Children:  MenuItems(m.GetChildren()),
	}
}

func MenuItems(ms []*v1_userv1.Menu) []types.MenuItem {
	out := make([]types.MenuItem, 0, len(ms))
	for _, m := range ms {
		out = append(out, MenuItem(m))
	}
	return out
}

// ============================================================
// Dept(递归)
// ============================================================

func DeptItem(d *v1_userv1.Dept) types.DeptItem {
	return types.DeptItem{
		DeptId:    d.GetDeptId(),
		ParentId:  d.GetParentId(),
		DeptName:  d.GetDeptName(),
		DeptType:  d.GetDeptType(),
		LeaderId:  d.GetLeaderId(),
		SortOrder: d.GetSortOrder(),
		Status:    d.GetStatus(),
		CreatedAt: Timestamp(d.GetCreatedAt()),
		Children:  DeptItems(d.GetChildren()),
	}
}

func DeptItems(ds []*v1_userv1.Dept) []types.DeptItem {
	out := make([]types.DeptItem, 0, len(ds))
	for _, d := range ds {
		out = append(out, DeptItem(d))
	}
	return out
}

// ============================================================
// Scope(数据权限)
// ============================================================

func ScopeItem(s *v1_userv1.Scope) types.ScopeItem {
	return types.ScopeItem{
		ScopeId:        s.GetScopeId(),
		RoleId:         s.GetRoleId(),
		ResourceType:   s.GetResourceType(),
		FieldName:      s.GetFieldName(),
		ConditionType:  s.GetConditionType(),
		ConditionValue: s.GetConditionValue(),
		Description:    s.GetDescription(),
		CreatedAt:      Timestamp(s.GetCreatedAt()),
	}
}

func ScopeItems(ss []*v1_userv1.Scope) []types.ScopeItem {
	out := make([]types.ScopeItem, 0, len(ss))
	for _, s := range ss {
		out = append(out, ScopeItem(s))
	}
	return out
}

// ============================================================
// User / UserProfile
// ============================================================

func UserItem(u *v1_userv1.User) types.UserItem {
	return types.UserItem{
		UserId:    u.GetUserId(),
		Username:  u.GetUsername(),
		Email:     u.GetEmail(),
		Phone:     u.GetPhone(),
		Status:    u.GetStatus(),
		CreatedAt: Timestamp(u.GetCreatedAt()),
		UpdatedAt: Timestamp(u.GetUpdatedAt()),
	}
}

func UserItems(us []*v1_userv1.User) []types.UserItem {
	out := make([]types.UserItem, 0, len(us))
	for _, u := range us {
		out = append(out, UserItem(u))
	}
	return out
}

// ============================================================
// Address
// ============================================================
//
// AddressItem 里**没有 user_id** —— 单体 response.AddressResp 也不返回它。
// 前端不需要(地址都是自己的),返回它反而多一个可被误用的字段。
//
// CreatedAt 用 Timestamp()(RFC3339 字符串)。
// **注意单体那边不是这样**:response.AddressResp 的 CreatedAt 是
// time.Time.String() 的输出(如 "2026-02-14 10:30:00 +0800 CST"),
// 与 RFC3339("2026-02-14T10:30:00Z")**格式不同**。
//
// 若前端直接把这个字段显示给用户,格式会变(从 "2026-02-14 10:30:00
// +0800 CST" 变成 "2026-02-14T02:30:00Z" —— 还可能差 8 小时)。
//
// **这是地址域的一处待确认差异。** 要对齐单体,得复用
// time.Time.String() 的格式;但那个格式含时区名,跨平台输出不稳定。
// 端到端验证时对比一次两个接口的 created_at 字段。
func AddressItem(a *v1_userv1.Address) types.AddressItem {
	return types.AddressItem{
		AddressId:     a.GetAddressId(),
		ReceiverName:  a.GetReceiverName(),
		ReceiverPhone: a.GetReceiverPhone(),
		Province:      a.GetProvince(),
		City:          a.GetCity(),
		District:      a.GetDistrict(),
		DetailAddress: a.GetDetailAddress(),
		PostalCode:    a.GetPostalCode(),
		IsDefault:     a.GetIsDefault(),
		AddressTag:    a.GetAddressTag(),
		CreatedAt:     Timestamp(a.GetCreatedAt()),
	}
}

func AddressItems(as []*v1_userv1.Address) []types.AddressItem {
	out := make([]types.AddressItem, 0, len(as))
	for _, a := range as {
		out = append(out, AddressItem(a))
	}
	return out
}
