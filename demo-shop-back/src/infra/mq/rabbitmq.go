// Package mq 消息队列层——RabbitMQ 连接管理与发件箱投递。
//
// **订单延迟队列拓扑已从本包移除**(consumer.go / order_delay.go /
// InitOrderDelayTopology 三处)。那套拓扑(order.dead.exchange /
// order.delay.queue / order.dead.queue)属于订单域,已随订单三表迁到
// trade-service —— 它在 infra/mq/client.go 里声明自己的 order.trade.* 拓扑。
//
// 两个服务**必须各用各的队列名**:共用会被抢先消费,而单体这侧的回调
// 取消的是已经停更的本地表,结果是消息被吞、订单永不超时取消。
// C3 的券对账踩过同一个坑(两服务共用锁名,互相把对方的值抹掉)。
//
// 本包现在只负责:**把本库(demo_shop)sys_outbox_message 里的消息投出去**。
// 那是单体自己的发件箱,与订单域无关。
package mq

import (
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQ RabbitMQ 连接封装。
//
// 保留 Conn 字段以便将来需要第二条 channel ——
// AMQP 的 channel 不是并发安全的,发布与消费不能共用一条。
type RabbitMQ struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
	closed  chan struct{}
}

// NewRabbitMQ 创建 RabbitMQ 连接并打开 Channel
// 接收值：dsn - 连接字符串,格式: amqp://user:pass@host:port/vhost
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

// Closed 投递器据此判断连接是否该收摊
func (r *RabbitMQ) Closed() <-chan struct{} {
	if r == nil {
		return nil
	}
	return r.closed
}

// Close 关闭 RabbitMQ 连接（先闭合退出信号，再 Channel 后 Connection）
func (r *RabbitMQ) Close() {
	if r == nil {
		return
	}
	select {
	case <-r.closed: // 已闭合过，防 double-close panic
	default:
		close(r.closed)
	}
	if r.Channel != nil {
		_ = r.Channel.Close()
	}
	if r.Conn != nil {
		_ = r.Conn.Close()
	}
}
