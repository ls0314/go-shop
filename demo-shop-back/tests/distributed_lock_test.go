package tests

import (
	"testing"
	"time"

	"demo-shop-back/src/task"

	"github.com/go-redsync/redsync/v4"
	goredisv9 "github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

// L1 互斥语义:他人持锁时 TryLock 失败,锁空闲时成功——多实例防重复执行的核心
func TestDistributedLock_MutualExclusion(t *testing.T) {
	rdb := newTestRedis(t)
	m := task.NewDistributedLockManager(rdb)

	// 用独立的 redsync 客户端模拟「另一个实例」先持锁
	rs := redsync.New(goredisv9.NewPool(rdb.Client))
	holder := rs.NewMutex("lock:task:test-l1", redsync.WithExpiry(time.Minute), redsync.WithTries(1))
	if err := holder.TryLock(); err != nil {
		t.Fatalf("模拟另一实例持锁失败: %v", err)
	}

	if release, ok := m.TryLock("test-l1", time.Minute); ok {
		t.Fatal("他人持锁时 TryLock 应返回 false")
	} else {
		_ = release
	}
	if _, err := holder.Unlock(); err != nil {
		t.Fatalf("释放模拟锁失败: %v", err)
	}

	release, ok := m.TryLock("test-l1", time.Minute)
	if !ok {
		t.Fatal("锁空闲时 TryLock 应成功")
	}
	release()
}

// L2 降级语义:无 Redis 时退化为进程内互斥,任务不停摆
func TestDistributedLock_DegradeToInProcess(t *testing.T) {
	m := task.NewDistributedLockManager(nil) // cache=nil → 走 mu 兜底
	release, ok := m.TryLock("test-l2", time.Minute)
	if !ok {
		t.Fatal("降级路径 TryLock 应成功")
	}
	if _, ok := m.TryLock("test-l2", time.Minute); ok {
		release()
		t.Fatal("进程内互斥:同锁名第二次应失败")
	}
	release()
	release2, ok := m.TryLock("test-l2", time.Minute)
	if !ok {
		t.Fatal("释放后应可重新获取")
	}
	release2()
}
