package mq

import (
	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/repository"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type TaskLock interface {
	TryLock(name string, expiry time.Duration) (release func(), ok bool)
}

const (
	outboxLockName     = "outbox:dispatch" // 实际 Redis 锁 = lock:task:outbox:dispatch
	outboxLockExpiry   = 5 * time.Minute   // 锁上界:远大于单轮投递耗时,到期自动释放防死锁
	outboxPollInterval = time.Second       // 轮询周期:延迟取消的及时性要求是分钟级,1s 绰绰有余
	outboxBatchSize    = 100               // 单轮上限:防止积压时单轮持锁过久
)

func StartOutboxDispatcher(r *RabbitMQ, repo *repository.OutboxMessageRepo, lock TaskLock) {
	if r == nil {
		log.Println("[outbox] MQ 未配置,投递器不启动")
		return
	}
	go func() {
		ticker := time.NewTicker(outboxPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-r.closed:
				return
			case <-ticker.C:
				dispatchOnce(r, repo, lock)
			}
		}
	}()
	log.Println("[INFO] outbox投递器已启动")
}

func dispatchOnce(r *RabbitMQ, repo *repository.OutboxMessageRepo, lock TaskLock) {
	if lock != nil {
		release, ok := lock.TryLock(outboxLockName, outboxLockExpiry)
		if !ok {
			metrics.OutboxDispatchTotal.WithLabelValues("lock_skipped").Inc()
			return
		}
		defer release()
	}

	msgs, err := repo.FetchPending(outboxBatchSize)
	if err != nil {
		log.Printf("[outbox] 拉取待投递消息失败: %v", err)
		return
	}

	metrics.OutboxPendingGauge.Set(float64(len(msgs)))

	for i := range msgs {
		m := &msgs[i]

		err := r.Channel.Publish(
			m.Exchange,
			m.RoutingKey,
			false, false,
			amqp091.Publishing{
				ContentType:  "text/plain",
				Body:         []byte(m.PayLoad),
				DeliveryMode: amqp091.Persistent, // 持久化
			},
		)
		if err != nil {
			if mErr := repo.MarkFailed(m.OutBoxId, int(m.RetryCount)); mErr != nil {
				log.Printf("[outbox] 标记退避失败 id=%d: %v", m.OutBoxId, mErr)
			}
			metrics.OutboxDispatchTotal.WithLabelValues("failed").Inc()
			log.Printf("[outbox] 投递失败进入退避: id=%d messageId=%s retry=%d err=%v",
				m.OutBoxId, m.MessageId, m.RetryCount, err)
			continue
		}
		if sErr := repo.MarkSent(m.OutBoxId); sErr != nil {
			log.Printf("[outbox] 标记 sent 失败(重启后重发,消费端幂等兜底): id=%d err=%v", m.OutBoxId, sErr)
		}
		metrics.OutboxDispatchTotal.WithLabelValues("sent").Inc()
	}
}
