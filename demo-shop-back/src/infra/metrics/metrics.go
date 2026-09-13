// Package metrics Prometheus 可观测体系(DS-A-22)
// 三层指标:HTTP 层(延迟/状态码/流量) + 业务层(领券成败/锁库存/超时取消/限流命中) + 基础设施层(DB 连接池水位)
//
// 标签规范(指标设计的头号事故点):
//   - route 标签一律用 c.FullPath() 路由模板(如 /receive/:id),绝不用实际 URL——
//     否则每个 templateId 生成一个新标签组合,指标基数(cardinality)爆炸打爆 Prometheus
//   - 任何用户级 ID 不得作为标签
//   - 使用默认 Registry:第三方库注册同名指标会 panic 冲突,当前依赖面可控
package metrics

import (
	"log"
	"net/http"
	"os"
	"time"

	"demo-shop-back/db"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// ---- HTTP 层(Four Golden Signals: 流量/延迟/错误) ----
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_server_requests_total",
		Help: "HTTP 请求总数",
	}, []string{"method", "route", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_duration_seconds",
		Help:    "HTTP 请求耗时分布",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "route"})

	// ---- 业务层(把 DS-A-19/21 的改造效果显性化) ----
	CouponReceiveTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "coupon_receive_total",
		Help: "优惠券领取结果计数",
	}, []string{"result"}) // success / sold_out / limit_exceeded / error

	CouponReceiveDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "coupon_receive_duration_seconds",
		Help:    "领券接口耗时(gate=闸门启用链路 / db_only=纯DB链路),对比可见降级占比",
		Buckets: []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"path"})

	StockLockTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "stock_lock_total",
		Help: "库存锁定结果计数",
	}, []string{"result"}) // success / not_enough / error

	OrderTimeoutCancelTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "order_timeout_cancel_total",
		Help: "MQ 订单超时取消处理结果计数",
	}, []string{"result"}) // processed / failed

	RateLimitRejectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_rejected_total",
		Help: "限流拒绝计数(按路由模板)",
	}, []string{"route"})

	// ---- 基础设施层(Four Golden Signals: 饱和度) ----
	dbPoolInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_in_use",
		Help: "当前使用中的 DB 连接数",
	})
	dbPoolOpen = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_open",
		Help: "已打开的 DB 连接数",
	})
	dbPoolWaitCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_wait_count",
		Help: "等待空闲连接的累计次数(持续增长=连接池饱和,DS-A-19 分析的雪崩前兆)",
	})
)

// StartMetricsServer 在独立内部端口暴露 /metrics。
// 为什么独立端口而非主端口挂载:与业务流量隔离、不被全局限流误伤、
// compose 内网才可达不暴露公网(指标含路径模板等部署信息)
func StartMetricsServer() {
	port := os.Getenv("DEMO_SHOP_METRICS_PORT")
	if port == "" {
		port = "9002"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("[INFO] Prometheus 指标端点启动: :%s/metrics", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Printf("[WARN] 指标服务退出: %v", err)
		}
	}()
}

// StartDBPoolSampler 采样 DB 连接池水位(30s 周期)。
// 饱和度指标必须轮询:Gauge 是 set 语义的瞬时值,没有事件可订阅,
// 只能后台定时读 sql.DBStats 打点——这正是它和 Counter/Histogram 的本质区别
func StartDBPoolSampler(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			sqlDB, err := db.GetDB().DB()
			if err != nil {
				continue
			}
			stats := sqlDB.Stats()
			dbPoolInUse.Set(float64(stats.InUse))
			dbPoolOpen.Set(float64(stats.OpenConnections))
			dbPoolWaitCount.Set(float64(stats.WaitCount))
		}
	}()
}
