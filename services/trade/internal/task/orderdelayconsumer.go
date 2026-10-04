package task

import (
	"context"
	"errors"
	"time"

	"demo-shop/services/trade/internal/dto/resp"
	"demo-shop/services/trade/internal/infra/mq"
	"demo-shop/services/trade/internal/model"

	"github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// 订单延迟取消的消费端
// ============================================================
//
// 消费 order.trade.dead.queue:消息在延迟队列里等满该单的剩余支付时限后
// 转入死信队列,消费者在这里执行超时取消。
//
// 它是超时取消链路的**精确那一路**;定时扫描是兜底那一路。
// 两路都在时才是双保险 —— 少了消费端,取消精度退化到扫描周期(1 分钟)。

// OrderDelayConsumerService 死信队列消费者
type OrderDelayConsumerService struct {
	consumer *mq.OrderDelayConsumer
	orderSvc orderCanceller
	// cancelTimeout 单次取消的超时。取消要调下游(释放库存逐 SKU 一次、
	// 退券一次),5s 与 tradeclient 的调用超时同一量级
	cancelTimeout time.Duration
}

// orderCanceller 消费端需要的唯一能力。
//
// 声明成窄接口而不是直接用 *service.OrderService:消费端只调一个方法,
// 依赖整个订单服务会让这个包与 service 包产生不必要的耦合面,
// 单测也要构造一大串依赖(product/coupon RPC、DB、雪花…)。
// 用窄接口后,测试只需实现这一个方法。
type orderCanceller interface {
	CancelOrderBySystem(ctx context.Context, orderId int64, operator string) (*resp.CancelOrderResp, error)
}

// 系统取消的操作人。写进 user_order_log.operator,与单体的 "系统" 逐字一致 ——
// 订单日志里这个值是对账判定"谁取消的"的依据
const systemOperator = "系统"

// NewOrderDelayConsumerService 创建消费者
func NewOrderDelayConsumerService(consumer *mq.OrderDelayConsumer, orderSvc orderCanceller) *OrderDelayConsumerService {
	return &OrderDelayConsumerService{
		consumer:      consumer,
		orderSvc:      orderSvc,
		cancelTimeout: 5 * time.Second,
	}
}

// Run 消费循环,由 ServiceGroup 在独立 goroutine 里跑
func (s *OrderDelayConsumerService) Run(ctx context.Context) {
	if s.consumer == nil {
		logx.Error("订单延迟取消消费端未启动:MQ 未配置。超时取消仅由定时扫描兜底")
		return
	}
	logx.Infof("订单延迟取消消费端已启动: 队列 %s", mq.QueueOrderTradeDead)
	defer func() { _ = s.consumer.Close() }()

	deliveries := s.consumer.Deliveries()
	if deliveries == nil {
		logx.Error("订单延迟取消消费端投递流为空,退出")
		return
	}

	for {
		select {
		case <-ctx.Done():
			logx.Info("订单延迟取消消费端已停止")
			return
		case <-s.consumer.Closed():
			logx.Error("订单延迟取消消费端因 MQ 连接断开而退出(队列中的消息不丢,重启后继续消费)")
			return
		case msg, ok := <-deliveries:
			if !ok {
				// 投递流被 broker 关闭(channel/连接断了)
				logx.Error("订单延迟取消消费端的投递流已关闭,退出")
				return
			}
			s.handleExpired(ctx, msg)
		}
	}
}

// handleExpired 处理一条超时消息。
//
// ============================================================
// Ack 语义(与单体最大的差别)
// ============================================================
//
// 单体是**无条件 Ack**(consumer.go:97)—— 取消失败的消息也被丢弃,
// 只能等定时扫描兜底。DS-A-23 的 G3 记的就是这条:它把 MQ 这一路
// 降级成了"尽力而为",双保险实际只有一重。
//
// 这里按失败性质分三种:
//
//	业务失败(已支付/已取消/不存在)  → Ack,**重投也没用**,重投只会空转
//	基础设施失败(DB 断连/下游不可用) → Nack 重投,**这才是重投能救的**
//	消息体损坏(解析不出来)          → Ack,重投一万次还是解析不出来
//
// 并对基础设施失败加了**重投次数上限**:超过就 Ack 放弃,交给定时扫描。
// 没有这个上限,一条永远失败的消息会在 broker 与消费者之间
// 无间隔地弹来弹去(hot loop),把 CPU 与日志打满 ——
// 这是 AMQP 手动确认最常见的生产事故。
func (s *OrderDelayConsumerService) handleExpired(ctx context.Context, msg amqp091.Delivery) {
	payload, err := mq.ParseOrderDelayPayload(string(msg.Body))
	if err != nil {
		// 解析不出来:重投无意义,确认丢弃并告警(单体同此处理)
		logx.Errorf("超时消息体损坏,丢弃: body=%q err=%v", string(msg.Body), err)
		_ = msg.Ack(false)
		return
	}
	if payload.OrderId <= 0 {
		logx.Errorf("超时消息缺少有效订单ID,丢弃: body=%q", string(msg.Body))
		_ = msg.Ack(false)
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, s.cancelTimeout)
	defer cancel()

	result, err := s.orderSvc.CancelOrderBySystem(callCtx, payload.OrderId, systemOperator)
	switch {
	case err == nil:
		// 取消失败但**补偿没做完**(库存没还、券没退):订单已取消是事实,
		// 但还欠着下游两笔账。不重投 —— 重投会被"已取消"挡回去,
		// 得到的还是同一个结果。留给对账收敛,这里只告警
		if result != nil && !result.Compensated {
			logx.Errorf("订单 %d 已超时取消,但库存释放或退券未完成(等待对账收敛): orderNo=%s",
				payload.OrderId, payload.OrderNo)
		} else {
			logx.Infof("订单超时取消成功: orderId=%d orderNo=%s", payload.OrderId, payload.OrderNo)
		}
		_ = msg.Ack(false)

	case isBusinessSkip(err):
		// 已被支付或被别的取消者处理:幂等命中,不是故障
		logx.Infof("订单超时取消跳过(状态已变,幂等命中): orderId=%d reason=%v", payload.OrderId, err)
		_ = msg.Ack(false)

	default:
		s.retryOrGiveUp(msg, payload, err)
	}
}

// retryOrGiveUp 基础设施失败:重投,但最多一次
func (s *OrderDelayConsumerService) retryOrGiveUp(msg amqp091.Delivery, payload *mq.OrderDelayCancelPayload, err error) {
	// msg.Redelivered 由 broker 置位:true 表示这条消息之前已经投递过至少一次。
	// 用它当"重投上限"的判据,而不是自己维护计数器 ——
	// 计数器要落库或放 Redis,而这里只需要"别无限弹"这一个语义。
	if msg.Redelivered {
		logx.Errorf("订单超时取消重投后仍失败,放弃并交给定时扫描: orderId=%d err=%v",
			payload.OrderId, err)
		_ = msg.Ack(false)
		return
	}

	logx.Errorf("订单超时取消失败(将重投一次): orderId=%d err=%v", payload.OrderId, err)
	// requeue=true:放回队列头。这条消息已过期,重投后会立即再次投递 ——
	// 所以上面那个"只重投一次"的上限是必需的,否则就是 hot loop
	if nErr := msg.Nack(false, true); nErr != nil {
		logx.Errorf("超时消息 Nack 失败(将靠 broker 在连接恢复后重投): orderId=%d err=%v",
			payload.OrderId, nErr)
	}
}

// isBusinessSkip 判定"重投也没用"的业务性失败。
//
// 判据是**订单状态类**错误:重投时订单状态不会变回 pending_pay,
// 所以重投只会得到同一个错误。基础设施类错误(DB 断连、下游不可用)
// 才是重投能救的,故不在这个集合里。
func isBusinessSkip(err error) bool {
	return errors.Is(err, model.ErrOrderAlreadyCancelled) ||
		errors.Is(err, model.ErrOrderCannotCancel) ||
		errors.Is(err, model.ErrOrderNotExist)
}
