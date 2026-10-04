package couponservicelogic

import (
	"errors"
	"time"

	"demo-shop/services/marketing/internal/infra/metrics"
	"demo-shop/services/marketing/internal/model"
)

// ============================================================
// 领券埋点(DS-A-22)
// ============================================================
//
// 原先打在单体的 CouponService 上,券域迁走后那层只能观察到"RPC 这一跳",
// 且 path 标签必然是假值。迁到服务端后两处都变准:
//
//   - 结果分类:sold_out / limit_exceeded 可能来自闸门也可能来自 DB 双防线
//     (业务语义相同),只有在服务端才能把它们与真正的基础设施故障分开;
//   - path 标签:"闸门判定"还是"降级直走 DB",本进程内就有答案
//     —— 单体那侧读不到闸门,只能恒填 db_only。
//
// 埋点集中在这一个文件、且由 ReceiveCoupon 的**单一出口**调用:
// 结果是四选一的封闭枚举,散在各 return 点写必然漏掉某个分支,
// 而漏掉的分支表现为"这个结果永远是 0",排查时很难想到是埋点没打。

// observeReceive 记录一次领券的指标。
//
// gateServed 是"闸门是否真的做出了判定",与"闸门是否启用"不是一回事:
// 启用了但本次异常降级、或回填失败后仍落到 DB,都算 db_fallback。
// 这正是单体那侧填不出来的区分(它读不到闸门)。
//
// path 的取值用 metrics 包里的常量而不是字面量:它们是面板与告警表达式
// 依赖的契约,拼错不会编译报错,只会表现为"这个面板永远没数据"。
func observeReceive(result string, gateServed bool, start time.Time) {
	path := metrics.PathDBFallback
	if gateServed {
		path = metrics.PathRedisGate
	}
	metrics.CouponReceiveTotal.WithLabelValues(result, path).Inc()
	metrics.CouponReceiveDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())
}

// receiveResultOf 把业务错误映射成指标的结果标签。
//
// 只认这两种**可预期的业务失败**:它们各自对应一个明确的容量语义
// (售罄 / 达到限领),是容量规划要看的两个数。其余一律归 error ——
// 把未知错误塞进 success 或某个业务标签,都会让面板说谎。
func receiveResultOf(err error) string {
	switch {
	case errors.Is(err, model.ErrCouponSoldOut):
		return "sold_out"
	case errors.Is(err, model.ErrCouponLimitExceeded):
		return "limit_exceeded"
	default:
		return "error"
	}
}
