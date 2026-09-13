package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ============================================================
// 令牌桶限流(DS-A-21):全局兜底 + 敏感路由按 IP 收紧
// ============================================================
// 算法选型(面试必考):
//   固定窗口 —— 有边界突刺(两个窗口临界处可放行 2 倍流量),只适合粗粒度兜底
//   漏桶     —— 恒速放行,拒绝一切突发,会误伤「页面一次加载并发 5 个请求」这类合法突发
//   令牌桶   —— 恒速 r 补令牌,桶容量 b 决定突发上限:允许受控突发,拒绝持续超速
// 选令牌桶,用 golang.org/x/time/rate(标准扩展库的成熟实现)——不手写算法,
// 但必须能口述原理:取不到令牌即拒绝,令牌按 r/s 恒速补充,桶空后只能等补充
//
// 两级设计:
//   全局兜底   所有请求共享一个桶 → 保护 DB/下游总容量(连接池耗尽的雪崩路径)
//   按 IP 收紧 登录/注册/领券等「动作型」路由 → 防爆破/防刷
//
// 多实例语义(要能讲清):进程内限流在多实例部署下 = 每实例各限一份,
// 全局总量变为 N×b;跨实例全局限流需要 Redis Lua(与 DS-A-19 预扣同款思路),
// 当前单体阶段进程内是准确的
//
// 前置依赖:ClientIP() 的正确性依赖 DS-A-20 的 SetTrustedProxies 修复——
// 否则 nginx 反代后所有请求的 IP 都是网关 IP,按 IP 限流会把全部用户打同一个桶

var (
	defaultGlobalLimiter *rate.Limiter
	globalOnce           sync.Once
	defaultIPSet         *ipLimiterSet
	ipOnce               sync.Once
)

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// GlobalRateLimit 全局兜底限流(挂载在 CORS 之后、所有业务路由之前)。
// 默认 200 r/s、桶 400:略高于正常峰值,只拦「持续超速」而非「合理突发」
func GlobalRateLimit() gin.HandlerFunc {
	globalOnce.Do(func() {
		defaultGlobalLimiter = rate.NewLimiter(
			rate.Limit(envInt("DEMO_SHOP_RATE_GLOBAL_RPS", 200)),
			envInt("DEMO_SHOP_RATE_GLOBAL_BURST", 400),
		)
	})
	return GlobalRateLimitWith(defaultGlobalLimiter)
}

// GlobalRateLimitWith 可注入桶的内部版本——单测用,保证确定性(不受环境变量影响)
func GlobalRateLimitWith(l *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isExempt(c) {
			c.Next()
			return
		}
		// 非阻塞 Allow:限流的目的是「快速拒绝」而非「排队」,
		// 排队(Wait)只会把上游的超载变成自己的延迟雪崩
		if !l.Allow() {
			reject(c)
			return
		}
		c.Next()
	}
}

// ipLimiterSet 按 IP 的独立令牌桶集合。
// 为什么是结构体而不是包级变量:单测需要多套互不污染的桶集合
// (httptest 无法按 IP 隔离包级状态),依赖注入让测试确定性成立
type ipLimiterSet struct {
	m       sync.Map // map[clientIP]*ipLimiterEntry
	r       rate.Limit
	b       int
	idleTTL time.Duration
}

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen int64 // unix nano,惰性更新——清理 goroutine 据此回收
}

func newIPRateLimiterSet(r rate.Limit, b int, idleTTL time.Duration) *ipLimiterSet {
	return &ipLimiterSet{r: r, b: b, idleTTL: idleTTL}
}

// PerIPRateLimit 按 IP 收紧限流(挂载在登录/注册/领券等动作型路由上)。
// 默认 5 r/s、桶 10:正常人手速不可能触达,脚本爆破则在第 11 个请求被拒
func PerIPRateLimit() gin.HandlerFunc {
	ipOnce.Do(func() {
		defaultIPSet = newIPRateLimiterSet(
			rate.Limit(envInt("DEMO_SHOP_RATE_IP_RPS", 5)),
			envInt("DEMO_SHOP_RATE_IP_BURST", 10),
			10*time.Minute,
		)
		go defaultIPSet.cleanupLoop() // 后台清理:sync.Map 只增不减是内存泄漏
	})
	return defaultIPSet.middleware()
}

func (s *ipLimiterSet) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isExempt(c) {
			c.Next()
			return
		}
		ip := c.ClientIP()
		now := time.Now().UnixNano()
		// LoadOrStore:惰性创建——大多数 IP 一生只来几次,
		// 为它们预先建桶是浪费;首次到访时创建即可
		v, _ := s.m.LoadOrStore(ip, &ipLimiterEntry{
			limiter:  rate.NewLimiter(s.r, s.b),
			lastSeen: now,
		})
		entry := v.(*ipLimiterEntry)
		entry.lastSeen = now
		if !entry.limiter.Allow() {
			reject(c)
			return
		}
		c.Next()
	}
}

// cleanupLoop 每分钟扫一遍,回收空闲超过 idleTTL 的 IP 桶。
// 不做清理的后果:每个到访过的新 IP 永久占一条内存,公网环境下是慢性内存泄漏
func (s *ipLimiterSet) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanupOnce()
	}
}

func (s *ipLimiterSet) cleanupOnce() {
	cutoff := time.Now().Add(-s.idleTTL).UnixNano()
	s.m.Range(func(key, value interface{}) bool {
		if value.(*ipLimiterEntry).lastSeen < cutoff {
			s.m.Delete(key)
		}
		return true
	})
}

// isExempt 探活豁免:healthz 不能被限流打死——否则依赖抖动时高频请求
// 吃光令牌,编排层把健康实例判死并重启,雪崩被放大成故障
func isExempt(c *gin.Context) bool {
	return c.FullPath() == "/api/v1/healthz"
}

// reject 429 + Retry-After:限流的礼貌语义是告诉客户端「什么时候可以再来」。
// 这里给保守的 1 秒(精确值需计算下一个令牌到期时刻,业务上秒级粒度足够;
// 响应体沿用全局统一格式,HTTP 状态码才是探针契约)
func reject(c *gin.Context) {
	c.Header("Retry-After", "1")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code":    429,
		"message": "请求过于频繁,请稍后重试",
		"data":    nil,
	})
}
