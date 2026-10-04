package mq

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// 订单域的 RabbitMQ 拓扑与投递
// ============================================================

const (
	// ExchangeOrderTradeDelay 延迟交换机(直连)。
	ExchangeOrderTradeDelay = "order.trade.delay.exchange"
	// QueueOrderTradeDelay 延迟队列。消息在此等待 TTL 到期后转入死信队列。
	QueueOrderTradeDelay = "order.trade.delay.queue"
	// ExchangeOrderTradeDead 死信交换机。
	ExchangeOrderTradeDead = "order.trade.dead.exchange"
	// QueueOrderTradeDead 死信队列,超时取消的消费者监听它。
	QueueOrderTradeDead = "order.trade.dead.queue"

	// RoutingKeyOrderTradeDelay 投递延迟消息用的路由键。
	//
	// 与建单时写进 outbox 的 routing_key **必须逐字一致** ——
	// 两边各自配置,漂移了不报错,只是消息永远进不了队列。
	RoutingKeyOrderTradeDelay = "order.trade.delay"
	// RoutingKeyOrderTradeDead 死信队列的绑定键。
	RoutingKeyOrderTradeDead = "order.trade.dead"

	// delayQueueMessageTTL 延迟队列的兜底 TTL。
	delayQueueMessageTTL = 15 * time.Minute
)

// Client 订单域的 MQ 客户端。
type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel

	// closed 关闭信号:投递器据此退出循环。
	// 用 channel 而不是 bool:多个 goroutine 都在等它
	closed chan struct{}
	once   sync.Once
}

// NewClient 建连并声明拓扑。
func NewClient(amqpURL string) (*Client, error) {
	if amqpURL == "" {
		return nil, errors.New("未配置 RabbitMQ 地址")
	}

	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("打开 channel 失败: %w", err)
	}

	c := &Client{conn: conn, channel: ch, closed: make(chan struct{})}
	if err := c.declareTopology(); err != nil {
		_ = c.Close()
		return nil, err
	}

	// 连接意外断开时关闭 closed,让投递器停下 —— 否则它会拿着
	// 已失效的 channel 反复失败,日志每分钟刷 60 条
	go c.watchClose()

	return c, nil
}

// declareTopology 声明交换机、队列与绑定。
func (c *Client) declareTopology() error {
	if err := c.channel.ExchangeDeclare(ExchangeOrderTradeDelay, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明延迟交换机 %s 失败: %w", ExchangeOrderTradeDelay, err)
	}
	if err := c.channel.ExchangeDeclare(ExchangeOrderTradeDead, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明死信交换机 %s 失败: %w", ExchangeOrderTradeDead, err)
	}

	// 延迟队列:TTL 到期后转入死信交换机
	if _, err := c.channel.QueueDeclare(QueueOrderTradeDelay, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    ExchangeOrderTradeDead,
		"x-dead-letter-routing-key": RoutingKeyOrderTradeDead,
		"x-message-ttl":             int32(delayQueueMessageTTL.Milliseconds()),
	}); err != nil {
		return fmt.Errorf("声明延迟队列 %s 失败: %w", QueueOrderTradeDelay, err)
	}
	if err := c.channel.QueueBind(QueueOrderTradeDelay, RoutingKeyOrderTradeDelay, ExchangeOrderTradeDelay, false, nil); err != nil {
		return fmt.Errorf("绑定延迟队列失败: %w", err)
	}

	// 死信队列:消费者在这里执行超时取消
	if _, err := c.channel.QueueDeclare(QueueOrderTradeDead, true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明死信队列 %s 失败: %w", QueueOrderTradeDead, err)
	}
	if err := c.channel.QueueBind(QueueOrderTradeDead, RoutingKeyOrderTradeDead, ExchangeOrderTradeDead, false, nil); err != nil {
		return fmt.Errorf("绑定死信队列失败: %w", err)
	}

	return nil
}

// watchClose 监听连接断开
func (c *Client) watchClose() {
	connClosed := c.conn.NotifyClose(make(chan *amqp091.Error, 1))
	select {
	case <-c.closed:
	case <-connClosed:
		logx.Error("订单域 MQ 连接已断开,outbox 投递器将停止(超时取消仍由定时扫描兜底)")
		c.once.Do(func() { close(c.closed) })
	}
}

// Closed 投递器据此判断是否该退出
func (c *Client) Closed() <-chan struct{} { return c.closed }

// Publish 投递一条消息。
func (c *Client) Publish(exchange, routingKey, body string, expiration time.Duration) error {
	if c == nil {
		return errors.New("MQ 未建连")
	}
	select {
	case <-c.closed:
		return errors.New("MQ 连接已关闭")
	default:
	}

	msg := amqp091.Publishing{
		ContentType: "application/json",
		Body:        []byte(body),
		// 持久化:延迟消息必须扛住 broker 重启
		DeliveryMode: amqp091.Persistent,
	}
	if expiration > 0 {
		msg.Expiration = fmt.Sprintf("%d", expiration.Milliseconds())
	}

	return c.channel.Publish(exchange, routingKey, false, false, msg)
}

// Close 关闭连接
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.once.Do(func() { close(c.closed) })
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
