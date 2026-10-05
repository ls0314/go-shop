package dept

import (
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
)

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
// 每个 logic 包各有一份,理由见 role 包 helpers.go 的说明。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errNothingToUpdate 局部更新里没有任何可更新的字段。
var errNothingToUpdate = errors.New("请求参数错误")

// 分页兜底(单体用 DefaultQuery("page","1") / ("pageSize","10"))。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// ============================================================
// 局部更新:types 的指针字段 → proto 的 FieldUpdate
// ============================================================
//
// **nil 指针 = 该字段不更新**。
//
// 部门域比其它域多一个**可空语义**的坑:leader_id(负责人)。
//
//	如果前端想"取消负责人",传 leader_id = 0 是最自然的做法,
//	但 0 与"未传"在非指针 int64 里分不开 —— 故 .api 里它是 *int64。
//	传 0 表示"置为 0"(即无负责人),不传表示不动。
//
// 服务端是否把 0 解释成"无负责人"需要核实;若它把 0 当成
// "用户ID=0"去查,会返回"用户不存在"。端到端验证时要试这一条。

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

// int64Field 构造 int64 类型的 FieldUpdate;ptr 为 nil 时返回 nil。
func int64Field(name string, ptr *int64) *v1_userv1.FieldUpdate {
	if ptr == nil {
		return nil
	}
	return &v1_userv1.FieldUpdate{
		Field: name,
		Value: &v1_userv1.FieldValue{
			Value: &v1_userv1.FieldValue_Int64Value{Int64Value: *ptr},
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
