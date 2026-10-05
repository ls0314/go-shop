package menu

import (
	"fmt"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

// ============================================================
// 分页兜底
// ============================================================

// 与单体 handler 的 DefaultQuery 口径一致(menu_handler.go 用
// c.DefaultQuery("page","1") / ("pageSize","10"))。
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
// maxPageSize 是**行为收紧**:单体没有上限。100 远高于菜单树的
// 实际规模(菜单通常几十条),故不会影响正常使用。
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
//
// 菜单域比其它域**多两个 bool 字段**(is_visible / is_cache),
// 而 bool 的"零值 false"与"未传"最容易混:
// 若用非指针 bool,客户端只想改 menu_name 时会把 is_visible
// 一起改成 false,侧边栏上那个菜单就消失了。
// 故 .api 里这两个字段必须是指针。

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

// ============================================================
// MetaInfo 的双向转换
// ============================================================
//
// .api 里 MetaInfo 是 interface{}(goctl 1.9.2 不认 any),
// 而 proto 里是 *structpb.Struct。
//
//	请求方向(CreateMenu):interface{} → *structpb.Struct
//	响应方向(converter): *structpb.Struct → map[string]interface{}
//
// 两个方向都要转,而它们的**失败模式不同**:
//
//	请求方向 可能失败 —— 客户端塞了 JSON 不支持的值
//	          (函数、channel、NaN、循环引用),那时必须回 400
//	响应方向 不该失败 —— AsMap() 对任何合法的 Struct 都成功,
//	          对 nil 返回 nil。见 converter 的实现。

// toProtoStruct types 的 interface{} → proto 的 *structpb.Struct。
//
// nil / 非对象值都返回 (nil, nil) 表示"不设置 meta_info" ——
// 因为 proto 的 Struct 只能表达 JSON 对象,传一个数组或标量
// 进去 structpb.NewStruct 会报错,而那更像"客户端传错了类型"。
//
// **空 map 返回 nil 而不是空 Struct**:proto3 里未设置的消息字段
// 与设置为空的字段序列化结果不同(前者整个字段不出现),
// 而 nil 更接近"客户端没传"的语义。
func toProtoStruct(v interface{}) (*structpb.Struct, error) {
	if v == nil {
		return nil, nil
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		// 不是对象 —— 报错而不是静默忽略。
		// meta_info 按约定是对象({"title": ..., "icon": ...}),
		// 传数组或标量说明客户端理解错了契约。
		return nil, fmt.Errorf("meta_info 必须是 JSON 对象,收到 %T", v)
	}
	if len(m) == 0 {
		return nil, nil
	}

	s, err := structpb.NewStruct(m)
	if err != nil {
		// 失败原因:值里有 JSON 不支持的类型(函数/channel/NaN/循环引用)。
		// 带上原错误,便于前端定位是哪个字段。
		return nil, fmt.Errorf("meta_info 无法转换为 JSON 对象: %w", err)
	}
	return s, nil
}

// ============================================================
// meta_info 在局部更新里**暂不支持** —— 这是一个已知缺口
// ============================================================
//
// proto 的 FieldValue 是个 oneof,只有三种类型:
//
//	string_value / int64_value / bool_value
//
// **没有结构化类型**。而 Menu.meta_info 是 google.protobuf.Struct
// (一个对象,如 {"title":"优惠券","icon":"ticket"})。
//
// 要把对象塞进 FieldValue,只能序列化成 JSON 字符串放进
// string_value,再由服务端反序列化。这是一个**跨服务的编码约定** ——
// 而我没有核到 user-service 的 UpdateMenu 是否按 JSON 解析
// 这个 string_value:
//
//	若它解析   → 约定成立,可以实现
//	若它当文本  → 存进去的是 JSON 源码,读出来时 AsMap() 会失败,
//	              meta_info 变成一堆转义字符
//
// 故 **UpdateMenu 不发送 meta_info 字段**(见 updatemenulogic.go)。
// 后果:通过接口改不了 meta_info(但创建时可以,因为
// CreateMenuReq 传的是完整的 Menu 消息,不需要过 FieldValue)。
//
// 要补上这一步,先核 user-service 的 UpdateMenu 实现,确认约定后
// 在这里加一个 json.Marshal + strField 的辅助函数。
