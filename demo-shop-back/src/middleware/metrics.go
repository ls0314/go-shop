package middleware

import (
	"strconv"
	"time"

	"demo-shop-back/src/infra/metrics"

	"github.com/gin-gonic/gin"
)

// HTTPMetrics HTTP 层指标中间件(DS-A-22)。
// 挂载在 gin.Default() 之后的第一个业务中间件——在 CORS 之前,
// 保证包括预检和 404 在内的所有请求都被计量
//
// 【关键设计】route 标签必须用 c.FullPath()(路由模板):
// 用实际 URL 的话,每个 templateId/订单号都会生成一个新标签组合,
// 时间序列数量随业务数据线性爆炸,Prometheus 内存会被打爆。
// FullPath 为空(404 等未匹配请求)用 "UNMATCHED" 兜底,同样只有一条序列
func HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "UNMATCHED"
		}
		status := strconv.Itoa(c.Writer.Status())
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		metrics.HTTPDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}
