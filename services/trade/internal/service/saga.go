package service

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// ============================================================
// Saga 骨架:正向步骤 + 逆序补偿
// ============================================================

type sagaStep struct {
	// name 步骤名
	name string
	// forward 正向动作
	forward func(ctx context.Context) error
	// compensate 补偿动作
	compensate func(ctx context.Context) error
}

type saga struct {
	steps     []sagaStep
	completed []sagaStep
}

func newSaga(steps ...sagaStep) *saga {
	// 预分配:步骤数固定，避免补偿路径再扩容
	return &saga{
		steps:     steps,
		completed: make([]sagaStep, 0, len(steps)),
	}
}

// run 执行 Saga。返回 failedStep 与 failedStatus 而不只是 error:
func (s *saga) run(ctx context.Context) (failedStep string, err error) {
	for _, step := range s.steps {
		if fErr := step.forward(ctx); fErr != nil {
			s.compensateAll(ctx)
			return step.name, fErr
		}
		s.completed = append(s.completed, step)
	}
	return "", nil
}

// compensateAll 逆序补偿所有已完成的步骤。
func (s *saga) compensateAll(ctx context.Context) {
	for i := len(s.completed) - 1; i >= 0; i-- {
		step := s.completed[i]
		if step.compensate == nil {
			continue
		}
		if cErr := step.compensate(ctx); cErr != nil {
			// 只告警不重试:补偿各自幂等,留给对账任务收敛。
			// 这里重试会把故障路径的耗时放大数倍,而收益很小
			logx.Errorf("Saga 补偿失败(等待对账收敛): step=%s err=%v", step.name, cErr)
		}
	}
	s.completed = s.completed[:0]
}
