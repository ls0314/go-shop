package order

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// ============================================================
// 时间字段
// ============================================================
//
// 与其它域一致:RFC3339 字符串。但**订单域有一个额外的坑** ——
// 单体这里用的是 time.Time 而**不是** timestamppb,两者序列化格式
// 不同:
//
//	timestamppb.MarshalJSON → "2026-02-14T10:30:00.000000000Z"(纳秒)
//	time.Time              → "2026-02-14T10:30:00Z"(RFC3339)
//
// 而单体的 tradeclient 是把 proto 的 Timestamp 转成 Go 的 time.Time
// (通过 .AsTime()),再由 gin 的 encoding/json 序列化 ——
// **走的是 time.Time 那条路**,即 RFC3339。
//
// 故这里用 RFC3339 与单体一致。这与其它域的结论相同,但订单域是
// 唯一一处需要留意"proto 的时间经过了 time.Time 中转"的地方。
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// parseTimeParam 把查询参数里的时间字符串解析成 Timestamp。
//
// ============================================================
// 为什么列表接口的时间筛选要 BFF 解析
// ============================================================
//
// 单体的 requset.GetOrderListReq 里是:
//
//	StartTime *time.Time `form:"start_time"`
//
// gin 绑定 query 时会用 encoding/json 的规则解析(?!!)—— 实际上
// gin 的 form 绑定走的是它自己的时间格式列表(包含 RFC3339 与
// "2006-01-02 15:04:05"),比 encoding/json 宽松。
//
// 而 proto 那边是 google.protobuf.Timestamp。故 BFF 必须把字符串
// 转成 Timestamp —— 这一步无法省略。
//
// **格式容错取哪个口径**:
//
//	只认 RFC3339                → 严格,但前端若传 "2026-02-14 10:30:00"
//	                              就会被拒(而单体接受)
//	RFC3339 + "2006-01-02 15:04:05" + "2006-01-02"
//	                            → 与 gin 的容错程度接近  ← 采用
//
// 空串返回 nil(表示不筛选),不报错 —— 不传时间范围是正常的。
func parseTimeParam(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, nil
	}
	// 按顺序试;gin 的 form 绑定也是"试一组格式"的思路
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t), nil
		}
	}
	return nil, errors.New("时间格式错误: " + s)
}

// ============================================================
// 分页兜底
// ============================================================

// 与单体 handler 的 DefaultQuery 口径一致(order_handler.go 用
// c.DefaultQuery("page", "1") / ("pageSize", "10"))。
//
// **注意这里是 pageSize(camelCase)** —— 而查询参数名本身是
// **snake_case** 的 page_size(requset.GetOrderListReq 的 form tag)。
// 两者不矛盾:DefaultQuery 的键名只是**兜底默认值的读取键**,
// 而真正的绑定走 struct 的 form tag。
//
// 故 BFF 这边:参数名用 page_size(与前端一致),默认值用 1/10
// (与单体一致)。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
//
// maxPageSize 是**行为收紧**(单体无上限)。订单列表页通常 10~50 条,
// 100 足够;要导出全量应当走专门的接口。
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
