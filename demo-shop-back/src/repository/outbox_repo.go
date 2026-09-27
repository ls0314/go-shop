package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"time"

	"gorm.io/gorm"
)

type OutboxMessageRepo struct {
	db *gorm.DB
}

func NewOutboxMessage() *OutboxMessageRepo {
	return &OutboxMessageRepo{
		db: db.DB,
	}
}

func (o *OutboxMessageRepo) WithTx(tx *gorm.DB) *OutboxMessageRepo {
	return &OutboxMessageRepo{
		db: tx,
	}
}

// ============================================================
// OutBoxMessage相关操作
// ============================================================

func (o *OutboxMessageRepo) CreateOutbox(outbox model.OutBoxMessage) error {
	return o.db.Create(&outbox).Error
}

func (o *OutboxMessageRepo) FetchPending(limit int64) ([]model.OutBoxMessage, error) {
	var msgs []model.OutBoxMessage
	err := o.db.Where("status = ? AND next_retry_at <= ?", model.OutboxPending, time.Now()).
		Order("id").Limit(int(limit)).Find(&msgs).Error
	return msgs, err
}

func (o *OutboxMessageRepo) MarkSent(id int64) error {
	return o.db.Model(&model.OutBoxMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  model.OutboxSent,
			"sent_at": time.Now(),
		}).Error
}

func (o *OutboxMessageRepo) MarkFailed(id int64, retryCount int) error {
	backoff := time.Duration(1<<min(retryCount, 8)) * time.Second
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	return o.db.Model(&model.OutBoxMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   retryCount + 1,
			"next_retry_at": time.Now().Add(backoff),
		}).Error
}
