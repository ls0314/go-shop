package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DecodeUpdateJSON 把 proto 传过来的 JSON 对象文本解成更新字段映射。
//
// 契约:局部更新类 RPC 用 `updates_json`(字段名 → 值)承载"改哪些字段",
// 空串表示无更新,返回空 map(而非 nil),避免下游 mapstructure 报错。
// 类目域与商品域两个 logic 包共用同一份实现 —— 语义必须一致,
// 否则同一种入参在两个域上会得到不同结果。
func DecodeUpdateJSON(updatesJSON string) (map[string]interface{}, error) {
	if updatesJSON == "" {
		return map[string]interface{}{}, nil
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal([]byte(updatesJSON), &out); err != nil {
		return nil, fmt.Errorf("解析更新字段失败: %w", err)
	}
	return out, nil
}

// JoinWarnings 把数据异常告警拼成一行,便于日志输出
func JoinWarnings(warnings []string) string {
	return strings.Join(warnings, "; ")
}
