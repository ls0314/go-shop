package lock

import (
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TaskLockManager 后台任务的跨实例互斥锁。
type TaskLockManager struct {
	store *redis.Redis
	local sync.Mutex
}

func NewTaskLockManager(store *redis.Redis) *TaskLockManager {
	return &TaskLockManager{store: store}
}

// TryLock 尝试取锁。返回 release 供 defer 调用;ok=false 表示别的实例在跑。
func (m *TaskLockManager) TryLock(name string, expiry time.Duration) (release func(), ok bool) {
	noop := func() {}

	if m == nil || m.store == nil {
		if m == nil {
			return noop, false
		}
		if !m.local.TryLock() {
			return noop, false
		}
		return func() { m.local.Unlock() }, true
	}

	lock := redis.NewRedisLock(m.store, "lock:task:"+name)
	// SetExpire 不可省:不设时 seconds=0,锁会以约 500ms 的 PX 写入 ——
	// 等价于没加锁(go-zero RedisLock 的坑)。
	lock.SetExpire(int(expiry.Seconds()))

	acquired, err := lock.Acquire()
	if err != nil {
		logx.Errorf("任务锁获取失败(本轮跳过): name=%s err=%v", name, err)
		return noop, false
	}
	if !acquired {
		return noop, false
	}

	return func() {
		if _, err := lock.Release(); err != nil {
			logx.Errorf("任务锁释放失败(等待自然过期): name=%s err=%v", name, err)
		}
	}, true
}
