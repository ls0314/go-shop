// Package mq 消息队列层——RabbitMQ 连接管理与订单延迟队列
//
// 拓扑结构：
//
//	Exchange: order.dead.exchange (direct)
//	  ├── Queue: order.dead.queue (消费者监听)
//	  └── Queue: order.delay.queue (TTL 2min → x-dead-letter → order.dead.queue)
//
// 消息流：CreateOrder → PublishOrderDelay → 2min TTL 过期 → DLX → 消费者 → CancelOrder
package mq

import (
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// 队列/交换机常量
const (
	ExchangeOrderDead   = "order.dead.exchange" // 死信交换机——接收 TTL 过期的消息
	QueueOrderDelay     = "order.delay.queue"   // 延迟队列——无消费者，TTL 过期自动转入死信
	QueueOrderDead      = "order.dead.queue"    // 死信队列——消费者监听，收到即执行取消逻辑
	RoutingKeyOrderDead = "order.dead"          // 死信路由键
)

// RabbitMQ RabbitMQ 连接封装
type RabbitMQ struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
	closed  chan struct{} // 关闭信号，通知消费者 goroutine 退出
}

// NewRabbitMQ 创建 RabbitMQ 连接并打开 Channel
// 接收值：dsn - 连接字符串，格式: amqp://user:pass@host:port/vhost
// 返回值：*RabbitMQ - 连接实例, error - 连接失败返回错误
func NewRabbitMQ(dsn string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
		closed:  make(chan struct{}),
	}, nil
}

// Close 关闭 RabbitMQ 连接（先 Channel 后 Connection）
func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}

// InitOrderDelayTopology 声明订单延迟队列拓扑（幂等，重复调用安全）
//
// 声明顺序：死信交换机 → 死信队列 → 绑定 → 延迟队列（带 TTL + DLX）
// 接收值：无
// 返回值：error - 任一声明失败时返回带上下文的错误（调用方应中止或告警，
//
//	否则后续消息会投递到不存在的队列而静默丢失）
func (r *RabbitMQ) InitOrderDelayTopology() error {
	// 死信交换机
	if err := r.Channel.ExchangeDeclare(ExchangeOrderDead, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明死信交换机 %s 失败: %w", ExchangeOrderDead, err)
	}

	// 死信队列——消费者监听
	if _, err := r.Channel.QueueDeclare(QueueOrderDead, true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明死信队列 %s 失败: %w", QueueOrderDead, err)
	}
	if err := r.Channel.QueueBind(QueueOrderDead, RoutingKeyOrderDead, ExchangeOrderDead, false, nil); err != nil {
		return fmt.Errorf("绑定死信队列 %s 到交换机 %s 失败: %w", QueueOrderDead, ExchangeOrderDead, err)
	}

	// 延迟队列——无消费者，TTL 过期自动转入死信交换机
	if _, err := r.Channel.QueueDeclare(QueueOrderDelay, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    ExchangeOrderDead,
		"x-dead-letter-routing-key": RoutingKeyOrderDead,
		"x-message-ttl":             int32(15 * time.Minute / time.Millisecond),
	}); err != nil {
		return fmt.Errorf("声明延迟队列 %s 失败: %w", QueueOrderDelay, err)
	}
	return nil
}
