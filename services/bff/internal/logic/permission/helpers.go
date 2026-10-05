package permission

import (
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
)

// errNoIdentity 取不到调用方身份。
//
// 只在一种情况下出现:**该路由没挂 Auth 中间件**(配置错误),
// 而不是用户的问题。故它是普通 error,经 response.Failure 落到 500 ——
// 让运维看见配置漏了,而不是静默返回空数据。
//
// Auth 中间件自己拒绝无令牌/坏令牌时回的是 401,走不到这里。
//
// **每个 logic 包各有一份同样的定义**(user / role / permission /
// menu / scope / dept / address)。Go 的包内标识符不跨包,而这些包
// 都要判身份。放进 converter 包不合适(那是映射层,不该有错误语义);
// 抽共享 errors 包又会让"哪个包用哪个错误"变得不明显。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// errNothingToUpdate 局部更新请求里没有任何可更新的字段。
//
// 见 role 包 helpers.go 的详细说明(两种情况:一个字段都没传,
// 或传的字段名都不认识 —— go-zero 的 httpx.Parse 会静默丢弃
// 未声明的字段)。
var errNothingToUpdate = errors.New("请求参数错误")

// 分页兜底,与单体 handler 里的 DefaultQuery 口径一致。
//
// **必须兜底**:.api 的分页字段是 optional int,不传时为 0,
// 而 proto 的 page/page_size 传 0 会让服务端行为不确定。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
//
// maxPageSize 是**行为收紧**:单体没有上限,传 pageSize=100000
// 会一次性捞全表。100 远高于任何页面的展示需求。
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
// **nil 指针 = 该字段不更新** —— 这是"局部更新"的关键。
// 若把 nil 当成"更新为空串",一次改权限名就会把描述清掉。

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
