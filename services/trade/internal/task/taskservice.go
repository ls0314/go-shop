package task

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/service"
)

// taskService 把"跑一个阻塞循环"的后台任务适配成 go-zero 的 service.Service。
type taskService struct {
	name string
	run  func(ctx context.Context)

	mu       sync.Mutex
	cancel   context.CancelFunc
	stopping bool
	started  sync.Once
}

func newTaskService(name string, run func(ctx context.Context)) *taskService {
	return &taskService{name: name, run: run}
}

func (t *taskService) Start() {
	t.started.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		t.mu.Lock()
		if t.stopping {
			t.mu.Unlock()
			cancel()
			return
		}
		t.cancel = cancel
		t.mu.Unlock()

		t.run(ctx)
	})
}

func (t *taskService) Stop() {
	t.mu.Lock()
	t.stopping = true
	cancel := t.cancel
	t.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// 编译期断言:必须满足 go-zero 的 service.Service 接口
var _ service.Service = (*taskService)(nil)

// NewLoopService 把一个"自带循环的阻塞函数"包装成 service.Service。
//
// 与 AsService 的差别:AsService 负责**起 ticker 并周期性调用**一个无状态函数;
// 本函数不碰周期 —— 被包装的函数自己管循环
// (outbox 投递器要在循环里同时等 ctx、等 MQ 断开、等 ticker 三个信号,
// 压成"无状态函数 + 外部 ticker"会丢掉"MQ 一断就退出"的能力)。
func NewLoopService(name string, run func(ctx context.Context)) service.Service {
	return newTaskService(name, run)
}
