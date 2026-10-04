package task

import (
	"context"
	"errors"
	"testing"

	"demo-shop/services/trade/internal/dto/resp"
	"demo-shop/services/trade/internal/infra/mq"
	"demo-shop/services/trade/internal/model"

	"github.com/rabbitmq/amqp091-go"
)

// ============================================================
// 消费端的 Ack 语义测试
// ============================================================
//
// 这组用例钉的是消费端**唯一的实质决策**:一条处理失败的消息,
// 该确认丢弃还是重投。
//
// 单体在这件事上是无条件 Ack(consumer.go:97)——取消失败的消息也被丢弃,
// 于是 MQ 这一路实际降级成"尽力而为",双保险只剩扫描那一重
// (DS-A-23 的 G3 记的就是这条)。改成区分对待之后,这个区分必须有测试守着:
// 写错一个分支的后果分别是"消息被静默丢弃"或"消息无限弹跳打满 CPU"。

// fakeAck 记录确认动作的 mock Acknowledger。
//
// amqp091 的 Delivery.Acknowledger 是接口,官方注释就写明
// "Applications can provide mock implementations in tests of Delivery handlers",
// 故这里不需要 broker 也能覆盖确认路径。
type fakeAck struct {
	acked    int
	nacked   int
	requeued bool
	lastTag  uint64
}

func (f *fakeAck) Ack(tag uint64, multiple bool) error {
	f.acked++
	f.lastTag = tag
	return nil
}

func (f *fakeAck) Nack(tag uint64, multiple, requeue bool) error {
	f.nacked++
	f.requeued = requeue
	f.lastTag = tag
	return nil
}

func (f *fakeAck) Reject(tag uint64, requeue bool) error {
	f.nacked++
	f.requeued = requeue
	return nil
}

// fakeCanceller 只实现 CancelOrderBySystem —— 窄接口的好处:
// 消费端只调一个方法,测试就不必构造 DB、RPC、雪花等一串依赖。
type fakeCanceller struct {
	result *resp.CancelOrderResp
	err    error
	// gotOrderId 记下实际传进去的订单号,用于验证负载解析正确
	gotOrderId int64
	gotOp      string
}

func (f *fakeCanceller) CancelOrderBySystem(ctx context.Context, orderId int64, operator string) (*resp.CancelOrderResp, error) {
	f.gotOrderId = orderId
	f.gotOp = operator
	return f.result, f.err
}

// delivery 造一条待处理投递
func delivery(body string, redelivered bool) (amqp091.Delivery, *fakeAck) {
	ack := &fakeAck{}
	return amqp091.Delivery{
		Acknowledger: ack,
		Body:         []byte(body),
		Redelivered:  redelivered,
		DeliveryTag:  7,
	}, ack
}

// okBody 一条格式正确的延迟取消消息
func okBody(orderId int64) string {
	return mq.BuildOrderDelayPayload(&model.UserOrder{
		OrderId:       orderId,
		OrderNo:       "DS1TEST",
		IdempotentKey: "1:test-key-0123456789abcdef",
	})
}

// TestConsumerAcksOnSuccess 取消成功 → Ack
func TestConsumerAcksOnSuccess(t *testing.T) {
	canceller := &fakeCanceller{result: &resp.CancelOrderResp{Compensated: true}}
	svc := NewOrderDelayConsumerService(nil, canceller)

	msg, ack := delivery(okBody(1001), false)
	svc.handleExpired(context.Background(), msg)

	if ack.nacked != 0 {
		t.Errorf("成功取消不该 Nack,得到 nacked=%d", ack.nacked)
	}
	if ack.acked != 1 {
		t.Errorf("成功取消应 Ack 一次,得到 acked=%d", ack.acked)
	}
	if canceller.gotOrderId != 1001 {
		t.Errorf("订单号应从负载解出: got=%d want=1001", canceller.gotOrderId)
	}
	if canceller.gotOp != systemOperator {
		t.Errorf("操作人应为 %q,得到 %q", systemOperator, canceller.gotOp)
	}
}

// TestConsumerAcksOnAlreadyCancelled 已被取消/已支付 → Ack,不重投
//
// 这是**幂等命中的正常路径**:MQ 至少一次投递必然会重复,
// 而重复时订单状态已经变了。重投只会拿到同一个错误,
// 于是消息在 broker 与消费者之间无间隔弹跳(hot loop)。
func TestConsumerAcksOnAlreadyCancelled(t *testing.T) {
	for _, bizErr := range []error{
		model.ErrOrderAlreadyCancelled,
		model.ErrOrderCannotCancel,
		model.ErrOrderNotExist,
	} {
		t.Run(bizErr.Error(), func(t *testing.T) {
			svc := NewOrderDelayConsumerService(nil, &fakeCanceller{err: bizErr})

			msg, ack := delivery(okBody(1002), false)
			svc.handleExpired(context.Background(), msg)

			if ack.nacked != 0 {
				t.Errorf("业务性失败不该 Nack(重投也一样),得到 nacked=%d", ack.nacked)
			}
			if ack.acked != 1 {
				t.Errorf("业务性失败应 Ack,得到 acked=%d", ack.acked)
			}
		})
	}
}

// TestConsumerNacksInfraFailureOnce 基础设施失败 → 重投一次
//
// **这是与单体最关键的差别**。单体在这里也是 Ack ——
// 于是 DB 断连期间的所有超时消息被直接丢掉,只能等扫描兜底。
// 重投才是这类失败的对策:DB 恢复了消息就能处理成功。
func TestConsumerNacksInfraFailureOnce(t *testing.T) {
	infraErr := errors.New("dial tcp 127.0.0.1:5432: connect: connection refused")
	svc := NewOrderDelayConsumerService(nil, &fakeCanceller{err: infraErr})

	msg, ack := delivery(okBody(1003), false)
	svc.handleExpired(context.Background(), msg)

	if ack.acked != 0 {
		t.Errorf("基础设施失败不该直接 Ack(那样消息就丢了),得到 acked=%d", ack.acked)
	}
	if ack.nacked != 1 {
		t.Fatalf("基础设施失败应 Nack 一次,得到 nacked=%d", ack.nacked)
	}
	// requeue 必须为 true:否则消息被丢弃而不是重新入队
	if !ack.requeued {
		t.Error("Nack 的 requeue 必须为 true,否则消息被丢弃而非重投")
	}
}

// TestConsumerGivesUpAfterRedelivery 重投后仍失败 → 放弃并 Ack
//
// **没有这条就会 hot loop**。requeue=true 把消息放回队列头,
// 而这条消息早已过期,于是会被立即再次投递 —— 失败一次就弹一次,
// 无间隔地刷 CPU 与日志。broker 的 Redelivered 标记就是为此存在的:
// 用它当"已重投过"的判据,不需要自己维护计数器。
func TestConsumerGivesUpAfterRedelivery(t *testing.T) {
	infraErr := errors.New("下游 product-service 不可用")
	svc := NewOrderDelayConsumerService(nil, &fakeCanceller{err: infraErr})

	msg, ack := delivery(okBody(1004), true) // Redelivered=true 表示已经重投过
	svc.handleExpired(context.Background(), msg)

	if ack.nacked != 0 {
		t.Errorf("已重投过仍失败不该再 Nack(会 hot loop),得到 nacked=%d", ack.nacked)
	}
	if ack.acked != 1 {
		t.Errorf("已重投过仍失败应 Ack 放弃(交给定时扫描),得到 acked=%d", ack.acked)
	}
}

// TestConsumerAcksCorruptPayload 负载解析不出来 → Ack,不重投
//
// 重投一万次还是解析不出来。单体也是这个处理(且是对的)。
func TestConsumerAcksCorruptPayload(t *testing.T) {
	for _, body := range []string{"", "not-a-number", `{"order_id":`, "0", "-5"} {
		t.Run(body, func(t *testing.T) {
			svc := NewOrderDelayConsumerService(nil, &fakeCanceller{})

			msg, ack := delivery(body, false)
			svc.handleExpired(context.Background(), msg)

			if ack.nacked != 0 {
				t.Errorf("坏消息不该 Nack,得到 nacked=%d", ack.nacked)
			}
			if ack.acked != 1 {
				t.Errorf("坏消息应 Ack 丢弃,得到 acked=%d", ack.acked)
			}
		})
	}
}

// TestConsumerAcksWhenCompensationPending 订单已取消但补偿没做完 → 仍 Ack
//
// 补偿(释放库存/退券)失败不回滚"订单已取消"这个事实,
// 而重投会被"已取消"挡回去、拿不到补偿 —— 所以重投没有意义,
// 留给对账收敛。这条用例把这个判断固定下来,免得后人"顺手"改成 Nack。
func TestConsumerAcksWhenCompensationPending(t *testing.T) {
	canceller := &fakeCanceller{result: &resp.CancelOrderResp{Compensated: false}}
	svc := NewOrderDelayConsumerService(nil, canceller)

	msg, ack := delivery(okBody(1005), false)
	svc.handleExpired(context.Background(), msg)

	if ack.nacked != 0 {
		t.Errorf("补偿未完成不该 Nack(重投拿不到补偿),得到 nacked=%d", ack.nacked)
	}
	if ack.acked != 1 {
		t.Errorf("补偿未完成仍应 Ack(订单确已取消),得到 acked=%d", ack.acked)
	}
}

// TestConsumerParsesLegacyPayload 兼容单体写的裸 orderId
//
// sys_outbox_message 是从单体迁过来的同一张表,里面可能还有 pending 的旧行。
// 读不出来那些订单就永远不超时取消,且不报错 —— 只是每轮都失败一次。
func TestConsumerParsesLegacyPayload(t *testing.T) {
	canceller := &fakeCanceller{result: &resp.CancelOrderResp{Compensated: true}}
	svc := NewOrderDelayConsumerService(nil, canceller)

	msg, ack := delivery("8899", false)
	svc.handleExpired(context.Background(), msg)

	if canceller.gotOrderId != 8899 {
		t.Errorf("旧格式负载应解出 orderId=8899,得到 %d", canceller.gotOrderId)
	}
	if ack.acked != 1 {
		t.Errorf("旧格式负载处理成功应 Ack,得到 acked=%d", ack.acked)
	}
}
