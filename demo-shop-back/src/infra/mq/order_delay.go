package mq

import (
	"strconv"

	"github.com/rabbitmq/amqp091-go"
)

// PublishOrderDelay 发送订单延迟消息——创建订单时调用
//
// 消息体为 orderId 的字符串形式，投递到延迟队列。
// TTL 过期后自动转入死信队列，由消费者执行取消逻辑。
//
// 接收值：orderId - 订单ID
// 返回值：error - RabbitMQ 不可用时返回错误（调用方应降级处理，不阻塞主流程）
func (r *RabbitMQ) PublishOrderDelay(orderId int64) error {
	body := strconv.FormatInt(orderId, 10)
	return r.Channel.Publish(
		"",              // 默认交换机，直接投递到指定队列
		QueueOrderDelay, // 路由键=队列名
		false, false,
		amqp091.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(body),
			DeliveryMode: amqp091.Persistent, // 持久化到磁盘，RabbitMQ 重启不丢
		},
	)
}
