package tests

import (
	"sync"
	"testing"

	"demo-shop-back/src/utils"
)

// S1 雪花 ID:2 goroutine 各取 10 万 → 全局无重复;单 goroutine 内严格单调递增
func TestSnowflake_NoDuplicateAndMonotonic(t *testing.T) {
	sf, err := utils.NewSnowflake(99)
	if err != nil {
		t.Fatalf("初始化雪花失败: %v", err)
	}
	const (
		goroutines = 2
		perG       = 100_000
	)

	start := make(chan struct{})
	results := make(chan []int64, goroutines)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ids := make([]int64, 0, perG)
			for i := 0; i < perG; i++ {
				id := sf.NextId()
				// 单调性只能在采集端验证:合并后的顺序受调度影响,不可断言
				if len(ids) > 0 && id <= ids[len(ids)-1] {
					t.Errorf("ID 未严格递增: prev=%d cur=%d", ids[len(ids)-1], id)
					return
				}
				ids = append(ids, id)
			}
			results <- ids
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	seen := make(map[int64]struct{}, goroutines*perG)
	for ids := range results {
		for _, id := range ids {
			if _, dup := seen[id]; dup {
				t.Fatalf("雪花 ID 重复: %d", id)
			}
			seen[id] = struct{}{}
		}
	}
}
