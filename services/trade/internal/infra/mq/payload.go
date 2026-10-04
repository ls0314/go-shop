package mq

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"demo-shop/services/trade/internal/model"
)

// errEmptyPayload 空负载
var errEmptyPayload = errors.New("延迟取消消息负载为空")

// ============================================================
// 订单延迟取消消息的负载格式
// ============================================================
//
// **这是投递器与消费者之间的契约**。两边分别实现,格式漂移了不会编译报错,
// 只会在运行时"消息收到了但解析不出来" —— 故格式定义放这里,
// 且投递器**只**用 BuildOrderDelayPayload 构造,不手拼字符串。
//
// 为什么是 JSON 而不是单体那样的单值(裸 orderId 字符串):
//   - 单体消费者需要的东西只有一个 orderId,因为取消是本地事务;
//   - trade 侧取消要跨服务(释放库存 + 退还券),消费者还需要知道
//     **用哪个幂等键**去调那些下游 —— 幂等键在订单表上,消费者要么
//     再查一次库,要么让消息带上。带上更省一次往返,也让消息自解释。
//
// order_id 保留:消费者无论如何都要读订单(判断状态、拿明细),
// 故它不是冗余字段,而是主键。

// OrderDelayCancelPayload 延迟取消消息的负载。
type OrderDelayCancelPayload struct {
	// OrderId 订单主键。消费者据此读订单
	OrderId int64 `json:"order_id"`
	// OrderNo 订单号,只用于日志与排障(消息里能直接看出是哪张单)
	OrderNo string `json:"order_no"`
	// IdempotentKey 订单上的幂等键。消费者释放库存/退还券时的幂等判据
	IdempotentKey string `json:"idempotent_key"`
	// ExpireAt 支付截止时间(RFC3339)。消费者据此复核"是否真的过期了" ——
	// MQ 的 TTL 只是近似(队列积压、broker 重启都会让它偏晚),
	// 真正的判据始终是订单表的这一列
	ExpireAt string `json:"expire_at"`
}

// BuildOrderDelayPayload 构造延迟取消消息的负载。
//
// 构造失败(理论上不会:字段都是基本类型)返回空串,
// 由调用方决定跳过还是兜底 —— 投递器会选择**标记失败并退避**,
// 而不是投一条解析不出来的消息出去。
func BuildOrderDelayPayload(order *model.UserOrder) string {
	if order == nil {
		return ""
	}
	p := OrderDelayCancelPayload{
		OrderId:       order.OrderId,
		OrderNo:       order.OrderNo,
		IdempotentKey: order.IdempotentKey,
	}
	if !order.ExpireAt.IsZero() {
		p.ExpireAt = order.ExpireAt.Format(time.RFC3339)
	}
	b, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseOrderDelayPayload 解析负载。消费者用。
//
// 容忍**裸 orderId 字符串**:单体时代写进 outbox 的就是那种格式,
// 而 trade_db 里可能还留着那时的行(status=pending 未投递)。
// 不兼容它们会让那些订单的超时取消永远发不出去。
func ParseOrderDelayPayload(body string) (*OrderDelayCancelPayload, error) {
	if body == "" {
		return nil, errEmptyPayload
	}
	// 先按新格式试
	if body[0] == '{' {
		var p OrderDelayCancelPayload
		if err := json.Unmarshal([]byte(body), &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
	// 兼容裸 orderId(单体格式)
	id, err := strconv.ParseInt(body, 10, 64)
	if err != nil {
		return nil, err
	}
	return &OrderDelayCancelPayload{OrderId: id}, nil
}
