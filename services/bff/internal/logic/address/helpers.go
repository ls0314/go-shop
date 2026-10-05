package address

import (
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
)

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
// 每个 logic 包各有一份,理由见 role 包 helpers.go 的说明。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errNothingToUpdate 局部更新里没有任何可更新的字段。
var errNothingToUpdate = errors.New("请求参数错误")

// ============================================================
// 地址域与其它域的**根本区别**:归属校验在服务端
// ============================================================
//
// RBAC 与档案域的 proto 入参只带 ID(如 GetRoleReq{role_id}),
// 服务端不知道请求来自谁 —— 故归属校验只能在 BFF 做
// (对比 JWT 的 user_id 与路径参数)。
//
// 而地址域的**每个 Req 都带 user_id**:
//
//	CreateAddressReq  { user_id, ... }
//	GetAddressReq     { user_id, address_id }
//	UpdateAddressReq  { user_id, address_id, updates }
//	DeleteAddressReq  { user_id, address_id }
//	SetDefaultAddressReq { user_id, address_id }
//
// 说明服务端的 Query 会带 `WHERE user_id = ? AND address_id = ?` ——
// **它自己保证归属**。
//
// 故 BFF 这边的做法是:
//
//	user_id **从 JWT 取**(middleware.UserID),不信请求体
//	address_id 从路径参数取
//
// 这样既不需要额外一次查询,也不需要 BFF 做归属比对 ——
// 服务端拿到 user_id 后会自然地把"别人的地址"过滤掉
// (查不到 → 返回业务错误)。
//
// **不要"顺手"在 BFF 加一次归属预检** —— 那会多一次 RPC,
// 且引入 TOCTOU。服务端的 WHERE 才是权威。

// ============================================================
// 局部更新:types 的指针字段 → proto 的 FieldUpdate
// ============================================================
//
// **nil 指针 = 该字段不更新**。
//
// 地址域的特别之处:is_default 是 *bool,而它关系到
// "切换默认地址"这个有副作用的操作 —— 服务端把某条设成默认时,
// 必须同时把同用户的其它地址取消默认(只能有一个默认)。
// 那是服务端的事务职责,BFF 不参与。

// strField 构造 string 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func strField(name string, ptr *string) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_StringValue{StringValue: *ptr},
		},
	}
}

// boolField 构造 bool 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func boolField(name string, ptr *bool) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_BoolValue{BoolValue: *ptr},
		},
	}
}

// compactFields 丢掉 nil 项(repeated 里塞 nil 会让服务端收到空更新)。
func compactFields(fields ...*v1_userv1.FieldUpdate) []*v1_userv1.FieldUpdate {
	out := make([]*v1_userv1.FieldUpdate, 0, len(fields))
	for _, f := range fields {
		if f != nil {
			out = append(out, f)
		}
	}
	return out
}
