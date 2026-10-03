package task

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/service"
)

// taskService 把"跑一个阻塞循环"的后台任务适配成 go-zero 的 service.Service。
//
// 为什么需要这层:service.Service 只有 Start() / Stop() 两个无参方法,
// 而任务需要 ctx(用于退出信号)与周期配置。这里持有 ctx 并在 Stop 时取消。
//
// Stop 可能先于 Start 到达(ServiceGroup 在启动失败时会 Stop 已注册的服务),
// 故用 stopping 标记记录"已经要求停过",Start 时立刻返回,不会跑起来。
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
