package task

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/service"
)

// taskService 把「需要 ctx 的任务函数」适配成 go-zero 的 service.Service。
//
// 为什么不直接让任务结构实现无参的 Start()/Stop():去掉 ctx 后任务就没法在
// 退出时中断当前轮次,ServiceGroup 的优雅退出会退化成"干等它跑完"。
// 这里由适配器持有 ctx 与 cancel,任务实现只管「按 ctx 干到被取消为止」。
type taskService struct {
	name string
	run  func(ctx context.Context)

	start sync.Once

	mu       sync.Mutex
	cancel   context.CancelFunc
	stopping bool // Stop 早于 Start 时置位,Start 起来后立刻取消
}

var _ service.Service = (*taskService)(nil)

func newTaskService(name string, run func(ctx context.Context)) *taskService {
	return &taskService{name: name, run: run}
}

// Start 阻塞运行任务;ServiceGroup 会把它放在自己的 goroutine 里。
func (t *taskService) Start() {
	t.start.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		// 与 Stop 在同一把锁下交接 cancel:ServiceGroup 的 doStop 与 doStart
		// 在不同 goroutine,Stop 可能先到 —— 那种情况下要立刻取消,
		// 否则任务会照跑一个完整周期才响应退出。
		t.mu.Lock()
		if t.stopping {
			cancel()
		}
		t.cancel = cancel
		t.mu.Unlock()

		t.run(ctx)
	})
}

// Stop 请求取消(幂等;可能先于 Start 被调用)
func (t *taskService) Stop() {
	t.mu.Lock()
	t.stopping = true
	cancel := t.cancel
	t.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}
