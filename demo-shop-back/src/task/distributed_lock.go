package task

import (
	"log"
	"sync"
	"time"

	"demo-shop-back/src/infra/cache"

	"github.com/go-redsync/redsync/v4"
	goredisv9 "github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

// DistributedLockManager 定时任务的跨实例互斥(DS-A-21)。
//
// 为什么不用手写 SETNX+EXPIRE:原子性要靠 SET NX EX 一条命令;释放锁必须
// 「校验 value 是自己的再删」(否则 A 的锁超时 → B 拿到 → A 跑完 DEL → 删掉了
// B 的锁 → C 又进来),而 GET+比较+DEL 不是原子的,又得靠 Lua——每个细节
// 都是坑。redsync 全部内置,且直接吃项目现有的 go-redis 连接池
//
// 为什么不用看门狗续期:Expiry(10min) 远大于单轮对账耗时(秒级),锁到期前
// 任务必然跑完;看门狗(定时 Extend)是「任务时长不可预算」场景的工具,
// 当前复杂度不值得——这是有意的取舍而非遗漏
//
// 锁的正确性边界(要能讲清):Expiry 到期+时钟漂移的极端组合下仍可能双持,
// Redlock 在工业界有著名争议(Kleppmann vs antirez)。对账任务幂等且低频,
// 双跑无害——锁只需做到「别重复浪费」,它是效率锁不是正确性锁
type DistributedLockManager struct {
	rs *redsync.Redsync
	mu sync.Mutex // Redis 不可用时的进程内兜底:退化为单机语义,任务不停摆
}

func NewDistributedLockManager(c *cache.RedisService) *DistributedLockManager {
	m := &DistributedLockManager{}
	if c != nil {
		m.rs = redsync.New(goredisv9.NewPool(c.Client))
	}
	return m
}

// TryLock 尝试获取任务锁。
// 返回 (release, true) = 拿到锁,任务执行完必须调用 release;
// 返回 (nil, false)  = 其他实例持有,本轮直接跳过——对账类任务幂等且低频,
//                      抢不到就放弃(WithTries(1)),不排队等待,避免任务积压
func (m *DistributedLockManager) TryLock(name string, expiry time.Duration) (release func(), ok bool) {
	if m.rs == nil {
		// 降级路径:无 Redis → 进程内互斥(单实例部署语义),任务不停摆。
		// 必须用 TryLock 而非 Lock:TryLock 的契约是「永不阻塞」,
		// 用阻塞锁会让"锁被占时的第二次尝试"永久挂起(测试 L2 抓出的死锁)
		if !m.mu.TryLock() {
			return nil, false
		}
		return func() { m.mu.Unlock() }, true
	}
	mutex := m.rs.NewMutex("lock:task:"+name,
		redsync.WithExpiry(expiry), // 到期自动释放防死锁;取值=任务耗时上界的数倍
		redsync.WithTries(1),       // 只试一次:对账任务宁跳过不排队
	)
	if err := mutex.TryLock(); err != nil {
		return nil, false
	}
	return func() {
		// 释放失败必须告警:典型原因是任务超时导致锁已过期易主——
		// 此刻出现了「双跑窗口」,对账任务幂等所以无害,但必须可见
		// (redsync 的 Unlock 返回 (bool, error),bool 表示是否真正释放)
		if _, err := mutex.Unlock(); err != nil {
			log.Printf("[WARN] 分布式锁释放失败(可能已过期易主): name=%s err=%v", name, err)
		}
	}, true
}
