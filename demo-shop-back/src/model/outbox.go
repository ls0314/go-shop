package model

import "time"

type OutBoxMessage struct {
	OutBoxId      int64      `gorm:"column:id;primary_key" json:"id"`
	MessageId     string     `gorm:"column:message_id" json:"message_id"`
	AggregateType string     `gorm:"column:aggregate_type" json:"aggregate_type"`
	AggregateId   string     `gorm:"column:aggregate_id" json:"aggregate_id"`
	EventType     string     `gorm:"column:event_type" json:"event_type"`
	Exchange      string     `gorm:"column:exchange" json:"exchange"`
	RoutingKey    string     `gorm:"column:routing_key" json:"routing_key"`
	PayLoad       string     `gorm:"column:payload" json:"payload"`
	Status        string     `gorm:"column:status" json:"status"`
	RetryCount    int64      `gorm:"column:retry_count" json:"retry_count"`
	NextRetryAt   time.Time  `gorm:"column:next_retry_at" json:"next_retry_at"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	SentAt        *time.Time `gorm:"column:sent_at" json:"sent_at"`
}

func (OutBoxMessage) TableName() string { return "sys_outbox_message" }
