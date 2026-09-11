package tests

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"errors"
	"sync"
	"testing"
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
