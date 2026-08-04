package mq

import (
	"demo-shop-back/src/model/response"
	"log"
	"strconv"

	"github.com/rabbitmq/amqp091-go"
)

// DefaultCanceller 全局订单取消回调——由 service.NewOrderService 通过 RegisterCanceller 注入
var DefaultCanceller OrderCanceller

// RegisterCanceller 注册订单取消回调实现
// 接收值：c - OrderCanceller 接口实现（通常为 *service.OrderService）
func RegisterCanceller(c OrderCanceller) {
	DefaultCanceller = c
}

// OrderCanceller 订单取消能力接口——由 service.OrderService 隐式实现
type OrderCanceller interface {
	CancelOrder(orderId, userId int64, userName string) (*response.OrderStatusResp, error)
}

// StartOrderConsumer 启动死信队列消费者（goroutine）——阻塞等待超时消息
//
// 消息处理流程：
//
//	收到消息 → 解析 orderId → 调用 CancelOrder → 状态机校验
//	  → pending_pay 则取消并释放库存
//	  → 非 pending_pay（已支付/已取消）则跳过
//
// 接收值：无——使用 DefaultCanceller 回调
// 返回值：error - 消费者注册失败时返回（goroutine 启动失败）
func (r *RabbitMQ) StartOrderConsumer() error {
	msgs, err := r.Channel.Consume(QueueOrderDead, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-r.closed:
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				r.handleOrderExpired(msg, DefaultCanceller)
			}
		}
	}()
	return nil
}

// handleOrderExpired 处理超时消息——解析 orderId 并调用取消逻辑
//
// 接收值：
//
//	msg       - AMQP 投递消息（body 为 orderId 字符串）
//	canceller - 取消回调接口
func (r *RabbitMQ) handleOrderExpired(msg amqp091.Delivery, canceller OrderCanceller) {
	// 防御：回调未注册时不能调用，否则 nil panic；消息直接确认丢弃并告警
	if canceller == nil {
		log.Printf("[MQ] 订单取消回调未注册，丢弃消息 orderId=%s", string(msg.Body))
		msg.Ack(false)
		return
	}

	orderId, err := strconv.ParseInt(string(msg.Body), 10, 64)
	if err != nil {
		// 消息体损坏无法解析，确认丢弃并告警，避免无限重投递
		log.Printf("[MQ] 解析订单ID失败，丢弃消息 body=%q err=%v", string(msg.Body), err)
		msg.Ack(false)
		return
	}

	// CancelOrder 内部有状态机校验：仅 pending_pay 可取消，其他状态自动跳过
	// 无论取消失败与否都 Ack，避免无限重试
	_, err = canceller.CancelOrder(orderId, 4, "系统")
	if err != nil {
		log.Printf("[MQ] 自动取消订单失败 orderId=%d err=%v", orderId, err)
	}
	msg.Ack(false)
}
