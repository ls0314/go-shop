package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"demo-shop-back/src/infra/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// M1 HTTP 指标:计数按 (method, route 模板, status) 归位,未匹配路由归入 UNMATCHED
func TestHTTPMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(HTTPMetrics())
	r.GET("/api/v1/demo/:id", func(c *gin.Context) { c.Status(200) })

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/demo/999", nil))
	}
	// 3 次请求都落在同一个路由模板标签上(而不是 3 个不同 URL 的序列)
	if n := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/demo/:id", "200")); n != 3 {
		t.Fatalf("路由模板计数应为 3, 实际 %v", n)
	}

	// 404 → UNMATCHED 单序列
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/nowhere", nil))
	if n := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "UNMATCHED", "404")); n != 1 {
		t.Fatalf("UNMATCHED 计数应为 1, 实际 %v", n)
	}
}
