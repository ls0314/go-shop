package payment

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// formatTimestamp timestamppb.Timestamp → RFC3339 字符串(nil 安全)。
//
// 与其它域同口径:单体的响应经由 time.Time 再序列化,故是 RFC3339。
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// parseTimeParam 把查询参数里的时间字符串解析成 Timestamp。
//
// 容错三种格式(RFC3339 / "2006-01-02 15:04:05" / "2006-01-02"),
// 因为 gin 的 form 绑定比 encoding/json 宽松 —— 只认 RFC3339 会拒掉
// 单体原本接受的输入。空串返回 nil(不筛选)。
func parseTimeParam(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t), nil
		}
	}
	return nil, errInvalidParam
}

// 分页兜底。
//
// **支付列表的默认 page_size 是 20**(单体 handler 注释:"默认20,最大100"),
// 与订单列表的 10 不同。
const (
	defaultPaymentPage     = 1
	defaultPaymentPageSize = 20
	maxPaymentPageSize     = 100
)

// normalizePaymentPage 兜底并夹紧支付列表的分页参数。
func normalizePaymentPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPaymentPage
	}
	if pageSize <= 0 {
		pageSize = defaultPaymentPageSize
	}
	if pageSize > maxPaymentPageSize {
		pageSize = maxPaymentPageSize
	}
	return page, pageSize
}
