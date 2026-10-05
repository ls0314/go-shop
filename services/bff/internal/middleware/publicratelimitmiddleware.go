package middleware

import (
	"demo-shop/services/bff/internal/config"
	"demo-shop/services/bff/internal/utils"
	"net/http"
	"sync"
	"time"

	"demo-shop/services/bff/internal/response"

	"golang.org/x/time/rate"
)

type PublicRateLimitMiddleware struct {
	m       sync.Map // map[string]*ipLimiterEntry
	r       rate.Limit
	b       int
	idleTTL time.Duration

	// cfg 传给 utils.ClientIP 做可信代理判定。
	cfg *config.Config
}

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen int64 // unix nano,惰性更新
}

// NewPublicRateLimitMiddleware 构造限流中间件。
func NewPublicRateLimitMiddleware(cfg *config.Config) *PublicRateLimitMiddleware {
	rps, burst := 5, 10

	if cfg != nil {
		if cfg.RateLimit.IPRPS > 0 {
			rps = cfg.RateLimit.IPRPS
		}
		if cfg.RateLimit.IPBurst > 0 {
			burst = cfg.RateLimit.IPBurst
		}
	}

	m := &PublicRateLimitMiddleware{
		r:       rate.Limit(rps),
		b:       burst,
		idleTTL: 10 * time.Minute,
		cfg:     cfg,
	}

	// 后台清理:不做就是慢性内存泄漏
	go m.cleanupLoop()
	return m
}

func (m *PublicRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := utils.ClientIP(r, m.cfg)
		now := time.Now().UnixNano()

		// LoadOrStore:惰性创建
		v, _ := m.m.LoadOrStore(ip, &ipLimiterEntry{
			limiter:  rate.NewLimiter(m.r, m.b),
			lastSeen: now,
		})
		entry := v.(*ipLimiterEntry)
		entry.lastSeen = now

		// 非阻塞 Allow:限流的目的是"快速拒绝"而非"排队
		if !entry.limiter.Allow() {
			// 与单体 reject 同一响应:429 + Retry-After + 统一信封
			response.TooManyRequests(w, response.RateLimitedMessage)
			return
		}
		next(w, r)
	}
}

// cleanupLoop 每分钟扫一遍,回收空闲超过 idleTTL 的 IP 桶
func (m *PublicRateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-m.idleTTL).UnixNano()
		m.m.Range(func(key, value any) bool {
			if value.(*ipLimiterEntry).lastSeen < cutoff {
				m.m.Delete(key)
			}
			return true
		})
	}
}
