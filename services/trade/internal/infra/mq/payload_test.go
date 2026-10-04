package mq

import (
	"encoding/json"
	"testing"
	"time"

	"demo-shop/services/trade/internal/model"
)

// TestPayloadRoundTrip 投递器写什么、消费者就得能读回什么。
//
// 这是投递器与消费者之间唯一的契约,而两边分别实现 ——
// 格式漂移不会编译报错,只表现为"消息收到了但解析不出来"。
func TestPayloadRoundTrip(t *testing.T) {
	expire := time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC)
	order := &model.UserOrder{
		OrderId:       88123,
		OrderNo:       "DS1ABCDEF",
		IdempotentKey: "42:9f8e7d6c-1234-5678-9abc-def012345678",
		ExpireAt:      expire,
	}

	body := BuildOrderDelayPayload(order)
	if body == "" {
		t.Fatal("负载构造失败:返回空串")
	}

	got, err := ParseOrderDelayPayload(body)
	if err != nil {
		t.Fatalf("解析自己构造的负载失败: %v", err)
	}
	if got.OrderId != order.OrderId {
		t.Errorf("OrderId 不一致: got=%d want=%d", got.OrderId, order.OrderId)
	}
	if got.OrderNo != order.OrderNo {
		t.Errorf("OrderNo 不一致: got=%q want=%q", got.OrderNo, order.OrderNo)
	}
	// 幂等键必须完整带过来:消费者要用它去调库存与券服务,
	// 少了它就只能再查一次库(而这正是把它放进消息要省掉的那次往返)
	if got.IdempotentKey != order.IdempotentKey {
		t.Errorf("IdempotentKey 不一致: got=%q want=%q", got.IdempotentKey, order.IdempotentKey)
	}
	// 时间要能往返:消费者据此复核"是否真的过期了"(TTL 只是近似)
	if got.ExpireAt != expire.Format(time.RFC3339) {
		t.Errorf("ExpireAt 不一致: got=%q want=%q", got.ExpireAt, expire.Format(time.RFC3339))
	}
}

// TestParseLegacyBareOrderId 必须能读单体时代写进 outbox 的裸 orderId。
//
// 为什么这条重要:trade_db.sys_outbox_message 是从单体迁过来的**同一张表**,
// 里面可能还留着 status=pending 的旧行(单体写的裸 orderId 字符串)。
// 不兼容它们,那些订单的超时取消就永远发不出去 —— 而且不报错,
// 只是投递器每轮都失败一次。
func TestParseLegacyBareOrderId(t *testing.T) {
	got, err := ParseOrderDelayPayload("88123")
	if err != nil {
		t.Fatalf("解析裸 orderId 失败: %v", err)
	}
	if got.OrderId != 88123 {
		t.Errorf("OrderId 解析错误: got=%d want=88123", got.OrderId)
	}
	// 旧格式没有幂等键 —— 消费者要能容忍,靠自己查库补上
	if got.IdempotentKey != "" {
		t.Errorf("旧格式不该有幂等键,得到 %q", got.IdempotentKey)
	}
}

// TestParsePayloadRejectsGarbage 空负载与垃圾输入必须报错,而不是静默返回零值。
//
// 静默返回零值会让消费者拿 OrderId=0 去取消订单 —— 那是一个不存在的单,
// 报错至少能让投递器把它标成退避并留下日志。
func TestParsePayloadRejectsGarbage(t *testing.T) {
	for _, body := range []string{"", "not-a-number", "{}extra", `{"order_id":`} {
		if _, err := ParseOrderDelayPayload(body); err == nil {
			t.Errorf("负载 %q 应当报错,却解析成功了", body)
		}
	}
}

// TestBuildPayloadNilOrder 空订单返回空串而不是 "null"。
//
// 返回 "null" 会被消费者解析成一个全零结构体,而投递器把空串视为
// "构造失败"并退避 —— 这是刻意的:宁可留下待对账的记录,
// 也不要发一条语义不明的消息出去。
func TestBuildPayloadNilOrder(t *testing.T) {
	if got := BuildOrderDelayPayload(nil); got != "" {
		t.Errorf("nil 订单应返回空串,得到 %q", got)
	}
}

// TestBuildPayloadOmitsZeroExpire 零值 expire_at 不出现在负载里。
//
// 零值时间是 0001-01-01,序列化出去会让消费者以为"这单早过期了"
// 而立即取消。宁可不带这个字段,让消费者以订单表为准。
func TestBuildPayloadOmitsZeroExpire(t *testing.T) {
	body := BuildOrderDelayPayload(&model.UserOrder{OrderId: 1})
	var raw map[string]any
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("负载不是合法 JSON: %v", err)
	}
	if v, ok := raw["expire_at"]; ok && v != "" {
		t.Errorf("零值 ExpireAt 不该被写进负载,得到 %v", v)
	}
}
