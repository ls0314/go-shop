package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// ============================================================
// Saga 骨架的补偿顺序测试
// ============================================================
//
// 为什么这个文件的用例是全套里最该有的:
//
// Saga 的正确性有两条**独立的**要求 ——
//
//	① 正向顺序:1→2→3,后一步依赖前一步的成果
//	② 补偿逆序:3 失败补 2、1;2 失败**只**补 1
//
// 第 ② 条最容易错,而且只在"中间某步失败"时才暴露 ——
// 那条路径平时几乎不被执行到。曾经真实存在过的写法是
// `for i := len(s.steps) - 1; ...`(遍历**计划**而不是**已完成**),
// 后果是:第 2 步失败时会给从没执行过的第 3 步调补偿,
// 而真正该补的第 1 步**反而被跳过**。
//
// 这组用例不依赖 DB、RPC、broker —— 纯函数桩,毫秒级。

// recorder 记录步骤的执行轨迹。
//
// 用"轨迹切片"而不是"每步一个 bool":顺序本身就是要断言的东西,
// 而分开的 bool 只能证明"都执行过",证明不了"按什么顺序执行"。
type recorder struct {
	trace []string
}

func (r *recorder) mark(s string) { r.trace = append(r.trace, s) }

// joined 便于失败时一眼看出实际顺序
func (r *recorder) joined() string { return strings.Join(r.trace, " → ") }

// step 造一个只记录轨迹、不报错的步骤
func (r *recorder) step(name string) sagaStep {
	return sagaStep{
		name:       name,
		forward:    func(ctx context.Context) error { r.mark("fwd:" + name); return nil },
		compensate: func(ctx context.Context) error { r.mark("cmp:" + name); return nil },
	}
}

// stepFailOn 造一个正向会失败的步骤
func (r *recorder) stepFailOn(name string, err error) sagaStep {
	return sagaStep{
		name:       name,
		forward:    func(ctx context.Context) error { r.mark("fwd:" + name); return err },
		compensate: func(ctx context.Context) error { r.mark("cmp:" + name); return nil },
	}
}

// TestSagaCompensatesOnlyCompletedSteps 只补偿**真正执行成功**的步骤。
//
// 这是整个骨架最核心的不变量。场景:三步,第 2 步失败。
// 期望轨迹:fwd:1 → fwd:2 → cmp:1
//
// 三个断言各自能单独抓住一种写错的方式:
//   - 出现 cmp:2 → 补偿了失败的当前步骤(它自己没成功,不需要补)
//   - 出现 cmp:3 → 补偿了从没执行过的步骤(遍历 steps 而非 completed)
//   - 没有 cmp:1 → 该补的漏了(同上,或提前 return)
func TestSagaCompensatesOnlyCompletedSteps(t *testing.T) {
	r := &recorder{}
	boom := errors.New("第 2 步失败")
	s := newSaga(
		r.step("券核销"),
		r.stepFailOn("锁库存", boom),
		r.step("建单"),
	)

	failedStep, err := s.run(context.Background())

	if !errors.Is(err, boom) {
		t.Fatalf("应原样返回正向失败的错误,得到 %v", err)
	}
	if failedStep != "锁库存" {
		t.Errorf("failedStep 应为失败的那一步: got=%q want=%q", failedStep, "锁库存")
	}

	want := []string{"fwd:券核销", "fwd:锁库存", "cmp:券核销"}
	if r.joined() != strings.Join(want, " → ") {
		t.Errorf("执行轨迹错误\n got: %s\nwant: %s", r.joined(), strings.Join(want, " → "))
	}
}

// TestSagaCompensatesInReverseOrder 多步失败时补偿必须**逆序**。
//
// 场景:四步,第 4 步失败 → 应补 3、2、1(不是 1、2、3)。
//
// 为什么顺序重要:正着补会经过"库存已还但券还占着"这类中间态,
// 而对账任务按幂等键扫描时,逆序补出来的中间态与正向的中间态形态一致,
// 不会误判成"走到了更后面的一步"。
func TestSagaCompensatesInReverseOrder(t *testing.T) {
	r := &recorder{}
	boom := errors.New("第 4 步失败")
	s := newSaga(
		r.step("A"),
		r.step("B"),
		r.step("C"),
		r.stepFailOn("D", boom),
	)

	if _, err := s.run(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("应返回 D 的失败,得到 %v", err)
	}

	want := "fwd:A → fwd:B → fwd:C → fwd:D → cmp:C → cmp:B → cmp:A"
	if r.joined() != want {
		t.Errorf("补偿必须逆序\n got: %s\nwant: %s", r.joined(), want)
	}
}

// TestSagaAllStepsSucceed 全部成功时**不做任何补偿**,且 failedStep 为空。
func TestSagaAllStepsSucceed(t *testing.T) {
	r := &recorder{}
	s := newSaga(r.step("A"), r.step("B"), r.step("C"))

	failedStep, err := s.run(context.Background())

	if err != nil {
		t.Fatalf("全成功不该有错误,得到 %v", err)
	}
	if failedStep != "" {
		t.Errorf("全成功时 failedStep 应为空串(调用方据此判「有没有失败过」),得到 %q", failedStep)
	}
	want := "fwd:A → fwd:B → fwd:C"
	if r.joined() != want {
		t.Errorf("全成功不该有补偿\n got: %s\nwant: %s", r.joined(), want)
	}
}

// TestSagaFirstStepFailsNothingToCompensate 第 1 步失败 → 无补偿可做。
//
// 边界:completed 为空。遍历 completed 的写法天然正确,
// 而"遍历 steps 再从 completed 里找"之类的写法容易在这里越界或空转。
func TestSagaFirstStepFailsNothingToCompensate(t *testing.T) {
	r := &recorder{}
	boom := errors.New("第 1 步就失败")
	s := newSaga(
		r.stepFailOn("券核销", boom),
		r.step("锁库存"),
		r.step("建单"),
	)

	failedStep, err := s.run(context.Background())

	if !errors.Is(err, boom) {
		t.Fatalf("应返回第 1 步的错误,得到 %v", err)
	}
	if failedStep != "券核销" {
		t.Errorf("failedStep 应为 %q,得到 %q", "券核销", failedStep)
	}
	if r.joined() != "fwd:券核销" {
		t.Errorf("第 1 步失败时不该有任何补偿,也不该继续执行后续步骤\n got: %s", r.joined())
	}
}

// TestSagaStopsForwardAtFirstFailure 正向遇到失败就**停下**,不继续后面的步骤。
//
// 这条容易被忽略:继续执行后面的步骤会产生**新的**副作用,
// 而那些步骤的成果没人负责清理(它们不在 completed 里)。
func TestSagaStopsForwardAtFirstFailure(t *testing.T) {
	r := &recorder{}
	s := newSaga(
		r.step("A"),
		r.stepFailOn("B", errors.New("boom")),
		r.step("C"),
		r.step("D"),
	)

	_, _ = s.run(context.Background())

	if strings.Contains(r.joined(), "fwd:C") || strings.Contains(r.joined(), "fwd:D") {
		t.Errorf("失败后不该继续执行后续步骤(C/D)\n got: %s", r.joined())
	}
}

// TestSagaSkipsNilCompensation 没有补偿动作的步骤被跳过,且**不破坏逆序**。
//
// 真实场景:建单是本地事务,失败即整体失败,它自己没有副作用要撤
// (compensate = nil)。它是最后一步,但为了覆盖"中间某步无补偿"
// 这里把它放在中间。
func TestSagaSkipsNilCompensation(t *testing.T) {
	r := &recorder{}
	noCompensate := sagaStep{
		name:       "建单",
		forward:    func(ctx context.Context) error { r.mark("fwd:建单"); return nil },
		compensate: nil, // 本地事务:无副作用需撤
	}
	s := newSaga(
		r.step("券核销"),
		noCompensate,
		r.stepFailOn("发货", errors.New("boom")),
	)

	_, _ = s.run(context.Background())

	want := "fwd:券核销 → fwd:建单 → fwd:发货 → cmp:券核销"
	if r.joined() != want {
		t.Errorf("无补偿的步骤应被跳过且不打乱逆序\n got: %s\nwant: %s", r.joined(), want)
	}
}

// TestSagaContinuesCompensatingAfterOneFails 一个补偿失败**不阻断**其余补偿。
//
// 这是刻意的设计:补偿失败只记日志,由对账收敛。
// 若一个补偿失败就中断,会留下"有的撤了、有的没撤"的更难收拾的状态。
func TestSagaContinuesCompensatingAfterOneFails(t *testing.T) {
	r := &recorder{}
	// B 的补偿会失败
	failingCompensate := sagaStep{
		name:       "B",
		forward:    func(ctx context.Context) error { r.mark("fwd:B"); return nil },
		compensate: func(ctx context.Context) error { r.mark("cmp:B"); return errors.New("下游不可用") },
	}
	s := newSaga(
		r.step("A"),
		failingCompensate,
		r.step("C"),
		r.stepFailOn("D", errors.New("boom")),
	)

	_, _ = s.run(context.Background())

	// A 的补偿必须在 B 失败之后仍然执行
	if !strings.Contains(r.joined(), "cmp:A") {
		t.Errorf("B 的补偿失败后 A 的补偿仍须执行\n got: %s", r.joined())
	}
	// C 在 B 之前补偿(逆序),也不能被跳过
	if !strings.Contains(r.joined(), "cmp:C") {
		t.Errorf("C 的补偿不该被跳过\n got: %s", r.joined())
	}
	want := "fwd:A → fwd:B → fwd:C → fwd:D → cmp:C → cmp:B → cmp:A"
	if r.joined() != want {
		t.Errorf("补偿轨迹错误\n got: %s\nwant: %s", r.joined(), want)
	}
}

// TestSagaClearsCompletedAfterCompensation 补偿后清空已完成列表。
//
// 为什么需要:同一个 saga 实例被复用时(比如调用方重试),
// 残留的 completed 会让下一轮凭空补偿上一轮的步骤。
// 清空后重复 run 是"从头再来"的语义。
func TestSagaClearsCompletedAfterCompensation(t *testing.T) {
	r := &recorder{}
	s := newSaga(
		r.step("A"),
		r.stepFailOn("B", errors.New("boom")),
	)

	_, _ = s.run(context.Background())
	if len(s.completed) != 0 {
		t.Fatalf("补偿后 completed 应被清空,仍剩 %d 项", len(s.completed))
	}

	// 再跑一轮:轨迹应是全新的,不带上一轮的残留
	r.trace = nil
	_, _ = s.run(context.Background())

	want := "fwd:A → fwd:B → cmp:A"
	if r.joined() != want {
		t.Errorf("第二轮不该带上一轮的残留\n got: %s\nwant: %s", r.joined(), want)
	}
}

// TestSagaPassesContextToSteps 步骤拿到的必须是同一个 ctx。
//
// 场景上重要:调用方用 ctx 传取消与超时,而补偿是在**失败路径**上跑的 ——
// 那时 ctx 很可能已经超时。步骤自己决定要不要尊重它,
// 但骨架不能替它们换一个 ctx(比如偷偷用 context.Background()),
// 那会让"请求已取消"这件事在补偿里凭空消失。
func TestSagaPassesContextToSteps(t *testing.T) {
	type ctxKey string
	const k ctxKey = "trace-id"
	ctx := context.WithValue(context.Background(), k, "abc123")

	var gotFwd, gotCmp any
	s := newSaga(
		sagaStep{
			name:       "A",
			forward:    func(c context.Context) error { gotFwd = c.Value(k); return nil },
			compensate: func(c context.Context) error { gotCmp = c.Value(k); return nil },
		},
		sagaStep{
			name:       "B",
			forward:    func(c context.Context) error { return errors.New("boom") },
			compensate: nil,
		},
	)

	_, _ = s.run(ctx)

	if gotFwd != "abc123" {
		t.Errorf("正向步骤应拿到调用方的 ctx: got=%v", gotFwd)
	}
	if gotCmp != "abc123" {
		t.Errorf("补偿步骤也应拿到同一个 ctx: got=%v", gotCmp)
	}
}
