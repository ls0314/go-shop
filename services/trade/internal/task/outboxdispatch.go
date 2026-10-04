package task

import (
	"context"
	"time"

	"demo-shop/services/trade/internal/infra/lock"
	"demo-shop/services/trade/internal/infra/mq"
	"demo-shop/services/trade/internal/model"
	"demo-shop/services/trade/internal/repository"

	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// outbox 投递器
// ============================================================

const (
	// outboxLockName 跨实例互斥。带 outbox 前缀以免与超时扫描的锁混淆
	outboxLockName = "outbox:dispatch"
	// outboxLockExpiry 锁上界:远大于单轮投递耗时,到期自动释放防死锁
	outboxLockExpiry = 5 * time.Minute
	// outboxBatchSize 单轮上限:防止积压时单轮持锁过久
	outboxBatchSize = 100
	// outboxBackoffBase 投递失败后的退避基数(第 n 次失败退避 n*基数)
	outboxBackoffBase = 5 * time.Second
	// outboxMaxBackoff 退避上限:超过后固定在 1 分钟重试,
	// 而不是无限拉长(否则一次长时间故障后消息要几小时才再试)
	outboxMaxBackoff = time.Minute
)

// OutboxDispatcher 发件箱投递器
type OutboxDispatcher struct {
	repo   *repository.OutboxRepo
	client *mq.Client
	locks  *lock.TaskLockManager
	// interval 轮询周期。延迟取消的及时性要求是分钟级,
	// 故秒级轮询绰绰有余
	interval time.Duration
}

// NewOutboxDispatcher 创建投递器
func NewOutboxDispatcher(repo *repository.OutboxRepo, client *mq.Client, locks *lock.TaskLockManager, interval time.Duration) *OutboxDispatcher {
	if interval <= 0 {
		interval = time.Second
	}
	return &OutboxDispatcher{repo: repo, client: client, locks: locks, interval: interval}
}

// Run 阻塞循环,由 ServiceGroup 在独立 goroutine 里跑
func (d *OutboxDispatcher) Run(ctx context.Context) {
	if d.client == nil {
		// MQ 未配置:不启动。**不是静默返回** —— 日志要说明
		// "超时取消现在只有扫描这一重保险",否则运维会以为双保险都在
		logx.Error("outbox 投递器未启动:RabbitMQ 未配置。超时取消仅由定时扫描兜底")
		return
	}
	logx.Infof("outbox 投递器已启动: 周期 %v, 批量 %d", d.interval, outboxBatchSize)

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logx.Info("outbox 投递器已停止")
			return
		case <-d.client.Closed():
			// MQ 连接断开:停止投递。继续跑只会每分钟刷几十条失败日志,
			// 而消息仍在库里 pending,重连后重启即可补投 —— 不丢
			logx.Error("outbox 投递器因 MQ 连接断开而退出(消息留在库中,重启后补投)")
			return
		case <-ticker.C:
			d.dispatchOnce()
		}
	}
}

// dispatchOnce 单轮投递。
func (d *OutboxDispatcher) dispatchOnce() {
	if d.locks != nil {
		release, ok := d.locks.TryLock(outboxLockName, outboxLockExpiry)
		if !ok {
			return
		}
		defer release()
	}

	msgs, err := d.repo.GetPendingList(outboxBatchSize)
	if err != nil {
		logx.Errorf("outbox 拉取待投递消息失败: %v", err)
		return
	}
	if len(msgs) == 0 {
		return
	}

	for _, m := range msgs {
		d.dispatchOne(m)
	}
}

// dispatchOne 投递一条消息
func (d *OutboxDispatcher) dispatchOne(m *model.OutBoxMessage) {
	// ---- 计算剩余延迟 ----
	now := time.Now()
	if m.NextRetryAt.After(now) {
		return
	}

	// 到点后才发:此时消息在延迟队列里应停留的时间 = 0(已过期),
	// 但给对方留一点余量避免时钟漂移导致"提前投递"。
	// 真正的过期判据是订单表的 expire_at,消费者的复核会兜住这点偏差
	expiration := time.Duration(0)

	if m.PayLoad == "" {
		// 负载为空:投出去也解析不了。标记失败并退避,
		// 让它在重试上限内被反复尝试(可能是构造时的瞬时问题)
		d.markRetry(m, "负载为空")
		return
	}

	err := d.client.Publish(m.Exchange, m.RoutingKey, m.PayLoad, expiration)
	if err != nil {
		d.markRetry(m, err.Error())
		return
	}

	if err := d.repo.MarkSent(m.Id); err != nil {
		// 消息已经发出去了,只是标记失败 —— 重启后会重发。
		// 消费端幂等,故这**不是**故障,只记日志
		logx.Errorf("outbox 标记已投递失败(重启后会重发,消费端幂等兜底): id=%d messageId=%s err=%v",
			m.Id, m.MessageId, err)
		return
	}
	logx.Infof("outbox 已投递: id=%d messageId=%s eventType=%s", m.Id, m.MessageId, m.EventType)
}

// markRetry 标记失败并按退避推迟下次尝试
func (d *OutboxDispatcher) markRetry(m *model.OutBoxMessage, reason string) {
	if m.RetryCount >= model.OutboxMaxRetry {
		// 超过重试上限:不再推迟(仍留在 pending 等人工/对账处理)。
		// 不改成 failed 状态是刻意的 —— failed 会让它从待投递集合里消失,
		// 而对账任务扫的正是"pending 且很久没动"的行
		logx.Errorf("outbox 投递超过重试上限,已停止重试(等待对账处理): id=%d messageId=%s retry=%d reason=%s",
			m.Id, m.MessageId, m.RetryCount, reason)
		return
	}

	next := time.Now().Add(backoff(m.RetryCount))
	if err := d.repo.MarkRetry(m.Id, next); err != nil {
		logx.Errorf("outbox 标记退避失败: id=%d err=%v", m.Id, err)
		return
	}
	logx.Errorf("outbox 投递失败进入退避: id=%d messageId=%s retry=%d next=%s reason=%s",
		m.Id, m.MessageId, m.RetryCount+1, next.Format(time.RFC3339), reason)
}

// backoff 线性退避:第 n 次失败等 n*基数,上限 1 分钟。
//
// 线性而非指数:延迟取消的容忍度是分钟级,
// 指数退避会让一条消息在几次失败后要等几十分钟,而那期间订单一直是待支付。
func backoff(retryCount int) time.Duration {
	d := time.Duration(retryCount+1) * outboxBackoffBase
	if d > outboxMaxBackoff {
		return outboxMaxBackoff
	}
	return d
}
