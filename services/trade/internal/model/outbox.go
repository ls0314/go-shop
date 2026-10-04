package model

import "time"

// OutBoxMessage 事务性发件箱结构体, 对应数据表sys_outbox_message。
//
// 订单域用它发**延迟取消消息**:下单时在同一个本地事务里写一条
// delay:{order_id},由投递器在 TTL 后发出。事务提交则消息必达 ——
// 这解决了"订单建了、消息没发"导致订单永远不超时取消的问题。
type OutBoxMessage struct {
	Id            int64      `gorm:"primaryKey;column:id" json:"id"`
	MessageId     string     `gorm:"column:message_id" json:"message_id"`
	AggregateType string     `gorm:"column:aggregate_type" json:"aggregate_type"`
	AggregateId   string     `gorm:"column:aggregate_id" json:"aggregate_id"`
	EventType     string     `gorm:"column:event_type" json:"event_type"`
	Exchange      string     `gorm:"column:exchange" json:"exchange"`
	RoutingKey    string     `gorm:"column:routing_key" json:"routing_key"`
	PayLoad       string     `gorm:"column:payload" json:"payload"`
	Status        string     `gorm:"column:status" json:"status"`
	RetryCount    int        `gorm:"column:retry_count" json:"retry_count"`
	NextRetryAt   time.Time  `gorm:"column:next_retry_at" json:"next_retry_at"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	SentAt        *time.Time `gorm:"column:sent_at" json:"sent_at"`
}

func (OutBoxMessage) TableName() string {
	return "sys_outbox_message"
}
