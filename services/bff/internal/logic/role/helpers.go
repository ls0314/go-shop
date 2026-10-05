package role

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
// **每个 logic 包各有一份同样的定义**(user / role / menu /
// permission / scope / dept / address)。Go 的包内标识符不跨包,
// 而这些包都要判身份。放进 converter 包不合适(那是映射层,
// 不该有错误语义);抽一个共享 errors 包又会让"哪个包用哪个错误"
// 变得不明显。七份一行定义换来的好处:每个域的 logic 读起来自洽。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// 分页兜底,与单体 handler 里的 DefaultQuery 口径一致。
//
// 单体 role_handler.go:
//
//	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
//	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
//
// **必须兜底**:.api 的分页字段是 optional int,不传时为 0,
// 而 proto 的 page/page_size 传 0 会让服务端行为不确定
// (返回全部或返回空页)。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
//
// maxPageSize 是**行为收紧**:单体没有上限,传 pageSize=100000
// 会一次性捞全表。100 远高于任何页面的展示需求;
// 若确实有"拉全部"的场景(导出),应当单独做一个不分页的接口。
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

// errNothingToUpdate 局部更新请求里没有任何可更新的字段。
//
// 两种情况都落到这里:
//
//	① 客户端一个字段都没传(空 body 或全是 null)
//	② 客户端传了字段但**名字都不认识** —— go-zero 的 httpx.Parse
//	   会静默丢弃未声明的字段(它不报"未知字段"错)
//
// 两种对客户端来说都是"这次调用没有意义",故归一成"请求参数错误"。
// 在 BFF 拦掉而不是转发:空 updates 到了服务端要么被当成"无变更"
// 要么报错,两种都不理想,且白白多一次 RPC。
var errNothingToUpdate = errors.New("请求参数错误")

// ============================================================
// 局部更新:types 的指针字段 → proto 的 FieldUpdate
// ============================================================
//
// proto 的 UpdateRoleReq 收的是 repeated FieldUpdate,每个是
// {field: "role_name", value: {string_value: "..."}}
// (见 api/proto/user/v1/common.proto,那里用 oneof 而非 optional)。
//
// **nil 指针 = 该字段不更新** —— 这是"局部更新"的关键。

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

// compactFields 丢掉 nil 项。
//
// repeated 字段里塞 nil 元素会让服务端收到 {field:"", value:nil}
// 这样的空更新 —— 它要么报错要么写坏数据。在 BFF 过滤掉,
// 比要求每个服务端都判空更可靠。
func compactFields(fields ...*v1_userv1.FieldUpdate) []*v1_userv1.FieldUpdate {
	out := make([]*v1_userv1.FieldUpdate, 0, len(fields))
	for _, f := range fields {
		if f != nil {
			out = append(out, f)
		}
	}
	return out
}
