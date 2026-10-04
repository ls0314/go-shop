package mq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// ============================================================
// 消费:独立的第二条 channel
// ============================================================
//
// **为什么不复用发布用的 channel**:
//
// AMQP 的 channel 不是并发安全的 —— 一个 channel 上同时做 Publish 与 Consume
// 会导致帧交错,表现为随机的 "unexpected frame" 或连接被 broker 关闭。
// outbox 投递器每秒 publish 一次,而消费者是长驻的,两者必然并发。
//
// 故开第二条 channel:**发布一条、消费一条**,各不相干。
// 连接仍然共用一条(RabbitMQ 的推荐用法:连接昂贵、channel 轻量)。

// OrderDelayConsumer 死信队列消费者。
//
// 单独一个类型而不是给 Client 加 Consume 方法:消费有自己的生命周期
// (自己那条 channel、自己的关闭信号),混进 Client 会让
// "Close 关的是哪条"变得含糊。
type OrderDelayConsumer struct {
	client  *Client
	channel *amqp091.Channel
	// msgs 投递流。channel 关闭时它会被 close
	msgs <-chan amqp091.Delivery
}

// NewOrderDelayConsumer 在死信队列上开启消费。
//
// prefetch 为 1:超时取消要调下游 RPC(释放库存 = 逐 SKU 一次),
// 预取多条只会让消息在内存里排队,而它们本就在队列里等 ——
// 拉太快反而挤占连接与下游。
func (c *Client) NewOrderDelayConsumer(queue string, prefetch int) (*OrderDelayConsumer, error) {
	if c == nil {
		return nil, fmt.Errorf("MQ 未建连")
	}
	select {
	case <-c.closed:
		return nil, fmt.Errorf("MQ 连接已关闭")
	default:
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("打开消费 channel 失败: %w", err)
	}

	if prefetch <= 0 {
		prefetch = 1
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("设置 prefetch 失败: %w", err)
	}

	// autoAck=false:手动确认。**这是必须的** ——
	// 自动确认在投递到客户端时就 Ack,处理失败的消息直接丢失,
	// 而"处理失败"恰恰是这套机制最需要兜住的情况。
	msgs, err := ch.Consume(
		queue,
		"",    // consumer tag 留空由 broker 生成
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("在队列 %s 上开启消费失败: %w", queue, err)
	}

	return &OrderDelayConsumer{client: c, channel: ch, msgs: msgs}, nil
}

// Deliveries 投递流。连接断开时它会被 broker 关闭。
func (d *OrderDelayConsumer) Deliveries() <-chan amqp091.Delivery {
	if d == nil {
		return nil
	}
	return d.msgs
}

// Closed 连接级关闭信号,供消费循环 select
func (d *OrderDelayConsumer) Closed() <-chan struct{} {
	if d == nil {
		return nil
	}
	return d.client.closed
}

// Close 关闭消费 channel(不动连接 —— 它可能是投递器在用)
func (d *OrderDelayConsumer) Close() error {
	if d == nil || d.channel == nil {
		return nil
	}
	return d.channel.Close()
}
