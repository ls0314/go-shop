package couponservicelogic

import (
	"testing"

	"demo-shop/services/marketing/internal/infra/metrics"
	"demo-shop/services/marketing/internal/model"
)

// TestReceiveResultOf 结果标签的分类。
//
// 这层分类**只在服务端做得准**:单体的 CouponService 只能看到 RPC 错误,
// 分不出"sold_out 是闸门挡的还是 DB 双防线挡的",也分不出
// "这是容量到顶(业务)还是下游挂了(基础设施)"。迁到服务端后能分,
// 于是必须有测试钉住 —— 分错的后果是面板说谎:
// 把基础设施故障计进 success 会让"领券成功率"虚高。
func TestReceiveResultOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"售罄", model.ErrCouponSoldOut, "sold_out"},
		{"达到限领", model.ErrCouponLimitExceeded, "limit_exceeded"},
		// 其余一律 error。**不能**归进 success 或任何一个业务标签 ——
		// 那些标签各自对应一个明确的容量语义,塞错了容量规划就会读错数
		{"模板不存在", model.ErrCouponTemplateNotExist, "error"},
		{"越权用券", model.ErrUseCouponNoPermission, "error"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := receiveResultOf(c.err); got != c.want {
				t.Errorf("receiveResultOf(%v) = %q, want %q", c.err, got, c.want)
			}
		})
	}
}

// TestPathLabelValues 钉住 path 标签的取值。
//
// 为什么值得单写一条:这两个字符串是**面板与告警表达式依赖的契约**,
// 拼错一个字母不会编译报错,只会表现为"这个面板永远没数据" ——
// 而那是最难往"标签写错了"上想的一类故障。
//
// 取值以 DS-A-26 §1.4 的设计表为准。注意单体实现里当初填的是
// "gate"/"db_only",与设计表的 redis_gate/db_fallback 不一致 ——
// 迁移时收敛到设计文档,这条用例就是那个决定的守卫。
func TestPathLabelValues(t *testing.T) {
	if metrics.PathRedisGate != "redis_gate" {
		t.Errorf("闸门链路的 path 标签应为 redis_gate(DS-A-26 §1.4),得到 %q", metrics.PathRedisGate)
	}
	if metrics.PathDBFallback != "db_fallback" {
		t.Errorf("降级链路的 path 标签应为 db_fallback(DS-A-26 §1.4),得到 %q", metrics.PathDBFallback)
	}
	// 两个取值必须不同,否则"降级占比"这个观察目标就失效了
	if metrics.PathRedisGate == metrics.PathDBFallback {
		t.Error("两个 path 取值不能相同 —— 那样就无法区分闸门链路与降级链路")
	}
}
