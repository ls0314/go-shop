package repository

import (
	"time"

	"demo-shop/services/trade/internal/model"

	"gorm.io/gorm"
)

// OutboxRepo 事务性发件箱数据层实例。
type OutboxRepo struct {
	DB *gorm.DB
}

// NewOutboxRepo 创建发件箱数据层实例
func NewOutboxRepo(conn *gorm.DB) *OutboxRepo {
	return &OutboxRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (o *OutboxRepo) WithTx(tx *gorm.DB) *OutboxRepo {
	return &OutboxRepo{DB: tx}
}

// CreateOutbox 写入待发消息。
func (o *OutboxRepo) CreateOutbox(msg *model.OutBoxMessage) error {
	return o.DB.Create(msg).Error
}

// GetPendingList 拉取待投递消息(投递器用),按 next_retry_at 升序。
func (o *OutboxRepo) GetPendingList(limit int) ([]*model.OutBoxMessage, error) {
	var list []*model.OutBoxMessage
	err := o.DB.Where("status = ? AND next_retry_at <= ?", model.OutboxPending, time.Now()).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// MarkSent 标记已投递
func (o *OutboxRepo) MarkSent(id int64) error {
	now := time.Now()
	return o.DB.Model(&model.OutBoxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  model.OutboxSent,
			"sent_at": &now,
		}).Error
}

// MarkRetry 投递失败:重试次数 +1,并按退避推迟下次尝试。
func (o *OutboxRepo) MarkRetry(id int64, nextRetryAt time.Time) error {
	return o.DB.Model(&model.OutBoxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("LEAST(retry_count + 1, ?)", model.OutboxMaxRetry),
			"next_retry_at": nextRetryAt,
		}).Error
}
