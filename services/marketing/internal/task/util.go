package task

import "strconv"

// parseInt 把闸门里的字符串计数转成 int64。
// 解析失败返回错误,由调用方视同"值不可信"→ 用 DB 值覆盖。
func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// itoa 整数转闸门计数用字符串
func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
