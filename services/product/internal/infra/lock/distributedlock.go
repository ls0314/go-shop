package lock

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TaskLockManager 定时任务的跨实例互斥。
//
// 从单体 src/task/distributed_lock.go 平移(DS-A-21 的锁语义不变),
// **实现从 redsync 换成 go-zero 自带的 RedisLock**,原因:
//   - go-zero 的 lockscript.lua 与单体注释要求的正确性完全一致:
//     获取时「GET 比较 → 是自己的就续期,否则 SET NX PX」整段在 Lua 里原子执行;
//     释放走 delscript.lua「校验 value 是自己的再删」,同样原子。
//     单体注释里点名的两个坑(过期易主后误删他人锁、GET+比较+DEL 非原子)都已被覆盖;
//   - 省掉一个依赖(redsync)与**第二个 go-redis 连接池** —— redsync 需要裸 go-redis 客户端,
//     而 go-zero 的 *redis.Redis 不暴露底层连接,得照 RedisConf 再造一个池。
//
// 锁的正确性边界(照搬单体结论,要能讲清):Expiry 到期 + 时钟漂移的极端组合下仍可能双持,
// Redlock 在工业界本就有争议(Kleppmann vs antirez)。对账任务幂等且低频,
// 双跑无害 —— 锁只需做到「别重复浪费」,它是效率锁不是正确性锁。
type TaskLockManager struct {
	store *redis.Redis
	// 无 Redis 时的进程内兜底:退化为单机语义,任务不停摆
	mu sync.Mutex
}

// NewTaskLockManager 构造锁管理器。store 为 nil 表示 Redis 未配置。
func NewTaskLockManager(store *redis.Redis) *TaskLockManager {
	return &TaskLockManager{store: store}
}

// TryLock 尝试获取任务锁。
//
// 返回 (release, true) = 拿到锁,任务执行完必须调用 release;
// 返回 (nil, false)    = 其他实例持有,本轮直接跳过 —— 对账类任务幂等且低频,
// 抢不到就放弃,不排队等待(redsync 的 WithTries(1) 等价物),避免任务积压。
//
// 与单体实现的差异:go-zero 的 RedisLock 没有「只试一次」选项,但它的 Acquire
// 本身就是单次 SET NX 语义(失败即返回 false),不存在重试循环 —— 契约等价。
func (m *TaskLockManager) TryLock(name string, expiry time.Duration) (release func(), ok bool) {
	if m.store == nil {
		// 降级路径:无 Redis → 进程内互斥(单实例部署语义),任务不停摆。
		// 必须用 TryLock 而非 Lock:TryLock 的契约是「永不阻塞」,
		// 用阻塞锁会让"锁被占时的第二次尝试"永久挂起(单体测试 L2 抓出的死锁)。
		if !m.mu.TryLock() {
			return nil, false
		}
		return func() { m.mu.Unlock() }, true
	}

	// 键前缀 lock:task: 与单体一致:Redis 是共享实例,靠前缀做键空间隔离
	rl := redis.NewRedisLock(m.store, "lock:task:"+name)
	// 必须显式 SetExpire:不设时 seconds=0,锁会以「500ms + 0」的极短 PX 写入,
	// 等于没有锁 —— 这是 go-zero RedisLock 的静默陷阱
	rl.SetExpire(int(expiry.Seconds()))

	acquired, err := rl.AcquireCtx(context.Background())
	if err != nil {
		// 拿不到锁与「Redis 报错」是两回事:前者跳过本轮,后者也跳过本轮,
		// 但必须留下日志 —— Redis 不可用会让所有实例都「抢不到锁」,任务集体停摆
		logx.Errorf("获取任务锁失败(本轮跳过): name=%s err=%v", name, err)
		return nil, false
	}
	if !acquired {
		return nil, false
	}

	return func() {
		released, err := rl.ReleaseCtx(context.Background())
		// 释放失败必须告警:典型原因是任务超时导致锁已过期易主 ——
		// 此刻出现了「双跑窗口」,对账任务幂等所以无害,但必须可见
		if err != nil {
			logx.Errorf("分布式锁释放失败(可能已过期易主): name=%s err=%v", name, err)
			return
		}
		if !released {
			logx.Errorf("分布式锁释放未生效(锁已不属于本实例,说明发生过期易主): name=%s", name)
		}
	}, true
}
