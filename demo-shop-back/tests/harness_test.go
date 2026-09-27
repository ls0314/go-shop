package tests

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
)

type concurrentResult struct {
	affected int64
	err      error
}

func runConcurrent(t *testing.T, n int, action func(i int) (int64, error)) (stats map[string]int, affectedSum int64) {
	t.Helper()

	start := make(chan struct{})
	results := make(chan concurrentResult, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			affected, err := action(i)
			results <- concurrentResult{affected, err}
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)

	stats = map[string]int{}
	for r := range results {
		affectedSum += r.affected
		switch {
		case r.err == nil:
			stats["success"]++
		case errors.Is(r.err, model.ErrCouponSoldOut):
			stats["sold_out"]++
		case errors.Is(r.err, model.ErrCouponLimitExceeded):
			stats["limit_exceeded"]++
		case errors.Is(r.err, model.ErrStockNotEnough):
			stats["not_enough"]++
		default:
			stats["other"]++
			t.Logf("[other #%d] %v", stats["other"], r.err)
		}
	}
	return stats, affectedSum
}

func queryInt64(t *testing.T, sql string, args ...interface{}) int64 {
	t.Helper()
	var v int64
	if err := db.DB.Raw(sql, args...).Scan(&v).Error; err != nil {
		t.Fatalf("查询失败[%s]: %v", sql, err)
	}
	return v
}

func queryString(t *testing.T, sql string, args ...interface{}) string {
	t.Helper()
	var v string
	if err := db.DB.Raw(sql, args...).Scan(&v).Error; err != nil {
		t.Fatalf("查询失败[%s]: %v", sql, err)
	}
	return v
}

// flushGateKeys 清除指定模板的闸门键(闸门用例前置,必须调用)。
// 必要性(DS-A-19 实测踩坑):
//  1. 测试库 TRUNCATE + 序列重置 → template_id 每轮复用,而闸门键 TTL 长达数天,
//     上一轮的售罄键(coupon:stock:{同id}=0)会让这一轮所有请求在闸门被直接拒绝
//  2. 测试与开发共用同一 Redis 实例(同 DB 0),开发库同 id 模板的键(含对账任务
//     写入的键)会与测试串台
//
// 用 SCAN 而非 KEYS:虽然此处量小,但保持生产惯例一致
func flushGateKeys(t *testing.T, rdb *cache.RedisService, templateID int64) {
	t.Helper()
	ctx := context.Background()
	patterns := []string{
		fmt.Sprintf("coupon:stock:%d", templateID),
		fmt.Sprintf("coupon:limit:%d", templateID),
		fmt.Sprintf("coupon:ucnt:%d:*", templateID),
	}
	for _, p := range patterns {
		var cursor uint64
		for {
			keys, next, err := rdb.Client.Scan(ctx, cursor, p, 100).Result()
			if err != nil {
				t.Fatalf("扫描闸门键失败: %v", err)
			}
			if len(keys) > 0 {
				if err := rdb.Del(ctx, keys...); err != nil {
					t.Fatalf("清除闸门键失败: %v", err)
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
}

func newTestRedis(t *testing.T) *cache.RedisService {
	t.Helper()
	addr := getenv("TEST_REDIS_ADDR", "localhost:6379")
	svc, err := cache.NewRedisService(&redis.Options{
		Addr:     addr,
		Password: getenv("TEST_REDIS_PASSWORD", "demoShop"),
		DB:       0,
	})
	if err != nil {
		t.Skipf("Redis 不可用(%s),跳过闸门用例: %v", addr, err)
	}
	t.Cleanup(func() { svc.Close() })
	return svc
}
