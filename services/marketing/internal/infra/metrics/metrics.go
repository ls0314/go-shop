// Package metrics 券域的 Prometheus 指标(DS-A-22)。
//
// **为什么领券埋点在服务端而不是单体那侧**:
//
// 原先 `coupon_receive_total` / `coupon_receive_duration_seconds` 打在单体的
// CouponService 上。券域迁到本服务之后,那个位置只能观察到"RPC 这一跳",
// 而且有一个标签必然是假值 —— 原设计的 `path` 标签要区分
// 「闸门服务(gate)」与「降级直走 DB(db_only)」,但闸门现在跑在**本进程**里,
// 单体根本看不到它有没有生效,强行保留只会得到一个恒为 db_only 的标签。
//
// 迁到这里之后两个好处:
//  1. 结果分类更准 —— sold_out / limit_exceeded 可能来自闸门也可能来自
//     DB 双防线,业务语义相同,但只有在服务端才能把它们与真正的
//     基础设施故障分开(单体只能看到 RPC 错误);
//  2. path 标签变成**真值** —— 这次到底是闸门判的还是降级到 DB,
//     本进程里就有答案。
//
// **指标名与标签名保持与单体一致**,不重命名:Prometheus 的时序是按
// 指标名 + 标签定位的,改名等于让已有面板与告警表达式全部失效。
// 迁移后 instance 标签会从单体实例变成 marketing-service 实例,
// 面板若按 instance 切片需要同步调整(那是部署侧的改动,不是契约)。
package metrics

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ============================================================
// 指标定义
// ============================================================
//
// **标签取值以 DS-A-22 的设计表为准,不沿用单体实现里的取值**。
// 两者当初就不一致:
//
//	DS-A-22 §指标表    coupon_receive_duration_seconds{path=redis_gate/db_fallback}
//	单体实现与 help 文本 path 实填 "gate"/"db_only"
//
// 既然埋点要迁到服务端、迁移本身就会让 instance 标签变化(面板无论如何
// 都要调整),这里就顺手把取值收敛到设计文档 —— 留着两套说法只会让
// "面板上的 redis_gate 为什么没数据"变成一个反复出现的疑问。

var (
	// CouponReceiveTotal 领券结果计数(DS-A-22 指标表)。
	//
	// 标签规范:只放**有限枚举**。领券失败的原因集合是封闭的
	// (售罄/超限/其它),不随用户或模板增长 ——
	// 反之把 templateId 之类放进来会让指标基数(cardinality)爆炸,
	// 这是 DS-A-22 开头专门点出的"头号事故点"。
	//
	// result 只列 success/sold_out/limit_exceeded 三个**可预期**结果;
	// 其余(基础设施故障)归 error —— 不把它塞进任何一个业务标签,
	// 否则面板会说谎(见 receivemetrics.go 的说明)。
	CouponReceiveTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "coupon_receive_total",
		Help: "优惠券领取结果计数(redis_gate=闸门判定 / db_fallback=降级直走DB)",
	}, []string{"result", "path"}) // result: success / sold_out / limit_exceeded / error

	// CouponReceiveDuration 领券耗时(DS-A-22 指标表:观察降级占比)。
	//
	// 分位数在 Prometheus 侧用 histogram_quantile 算,不在这里预先算 ——
	// 预聚合分位数不可加,跨实例相加是错的。
	CouponReceiveDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "coupon_receive_duration_seconds",
		Help:    "领券耗时(redis_gate=闸门服务链路 / db_fallback=降级直走DB链路),对比可见降级占比",
		Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"path"})
)

// PathRedisGate / PathDBFallback 是 path 标签的两个取值(DS-A-22 用语)。
//
// 做成常量而不是散落的字面量:它们是**面板与告警表达式依赖的契约**,
// 拼错一个字母不会编译报错,只会表现为"这个面板永远没数据"。
const (
	// PathRedisGate 本次由 Redis 闸门做出了判定
	PathRedisGate = "redis_gate"
	// PathDBFallback 本次未走闸门(未启用 / 闸门异常降级 / 回填失败后直走 DB)
	PathDBFallback = "db_fallback"
)

// StartMetricsServer 在独立内部端口暴露 /metrics。
//
// 与单体同一做法与同一理由:**独立端口而非主端口挂载** ——
// 与业务流量隔离、不被全局限流误伤、compose 内网才可达
// (指标含路径模板等部署信息,不该暴露公网)。
//
// 端口从配置读,默认 9008(单体 9002;9006 是本服务的 RPC,故不复用)。
func StartMetricsServer(port string) {
	if port == "" {
		port = os.Getenv("DEMO_SHOP_MARKETING_METRICS_PORT")
	}
	if port == "" {
		port = "9008"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("[INFO] marketing-service Prometheus 指标端点: :%s/metrics", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			// 指标服务起不来不该拖垮业务:它是观测手段,不是依赖
			log.Printf("[WARN] marketing-service 指标服务退出: %v", err)
		}
	}()
}
