package scope

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
// 数据权限域的特别之处:它可改的字段里有一个**语义敏感**的
// field_name 与 condition_value —— 改它们等于改了"能看哪些数据"。
// 但那是业务风险,不是编码问题,故这里照常构造。

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
