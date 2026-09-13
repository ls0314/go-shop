package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func newTestEngine(handler gin.HandlerFunc, routes ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/healthz", func(c *gin.Context) { c.Status(200) })
	for _, p := range routes {
		r.GET(p, handler, func(c *gin.Context) { c.Status(200) })
	}
	return r
}

func do(r *gin.Engine, path, ip string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = ip + ":12345"
	r.ServeHTTP(w, req)
	return w
}

// R1 全局桶:突发 3 个通过,第 4 个被拒(桶容量即突发上限)+ Retry-After
func TestGlobalRateLimit_Burst(t *testing.T) {
	r := newTestEngine(GlobalRateLimitWith(rate.NewLimiter(1, 3)), "/t")
	for i := 0; i < 3; i++ {
		if w := do(r, "/t", "10.0.0.1"); w.Code != 200 {
			t.Fatalf("前 3 个请求应通过, 第 %d 个得到 %d", i+1, w.Code)
		}
	}
	w := do(r, "/t", "10.0.0.1")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("第 4 个请求应 429, 实际 %d", w.Code)
	}
	if w.Header().Get("Retry-After") != "1" {
		t.Fatal("429 必须携带 Retry-After")
	}
}

// R2 按 IP 隔离:一个 IP 打爆不影响另一个 IP
func TestPerIPRateLimit_Isolation(t *testing.T) {
	set := newIPRateLimiterSet(1, 2, time.Minute)
	r := newTestEngine(set.middleware(), "/t")

	do(r, "/t", "10.0.0.1")
	do(r, "/t", "10.0.0.1")
	if w := do(r, "/t", "10.0.0.1"); w.Code != 429 {
		t.Fatal("IP1 第 3 个请求应 429")
	}
	if w := do(r, "/t", "10.0.0.2"); w.Code != 200 {
		t.Fatalf("IP2 应不受 IP1 影响, 实际 %d", w.Code)
	}
}

// R3 探活豁免:healthz 永不被限流打死(否则编排层误判实例死亡,雪崩放大)
func TestHealthzExempt(t *testing.T) {
	r := newTestEngine(GlobalRateLimitWith(rate.NewLimiter(1, 1)), "/t")
	for i := 0; i < 10; i++ {
		if w := do(r, "/api/v1/healthz", "10.0.0.1"); w.Code != 200 {
			t.Fatalf("healthz 第 %d 次被限流(%d)", i+1, w.Code)
		}
	}
}

// R4 清理:僵尸桶被回收,活跃桶保留(防 sync.Map 内存泄漏)
func TestIPSetCleanup(t *testing.T) {
	set := newIPRateLimiterSet(1, 10, time.Minute)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t", set.middleware(), func(c *gin.Context) { c.Status(200) })
	do(r, "/t", "10.0.0.1") // 活跃条目
	set.m.Store("10.9.9.9", &ipLimiterEntry{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Add(-time.Hour).UnixNano(), // 人造僵尸
	})
	set.cleanupOnce()
	if _, ok := set.m.Load("10.9.9.9"); ok {
		t.Fatal("僵尸条目应被回收")
	}
	if _, ok := set.m.Load("10.0.0.1"); !ok {
		t.Fatal("活跃条目不应被回收")
	}
}
