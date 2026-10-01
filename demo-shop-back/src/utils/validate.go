package utils

import "regexp"

// ValidatePhone 手机号格式校验
// 规则：中国大陆11位手机号
// 接收值：phone - 手机号
// 返回值：bool - 校验结果
func ValidatePhone(phone string) bool {
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(phone)
	return reg
}
