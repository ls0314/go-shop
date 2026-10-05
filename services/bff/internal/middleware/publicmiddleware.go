// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import "net/http"

// ============================================================
// 免鉴权的完整清单(改这里之前先核一遍)
// ============================================================
//
// 本步的 bff.api 里只有两条用 Public:
//
//	POST /api/v1/user/refresh          刷新令牌(由 user-service 自行验签)
//	POST /api/v1/pay/callback/mock     支付渠道回调(渠道不带我们的令牌)
//
// 后续批次还会有:
//
//	GET  /api/v1/healthz               探活(探针不该需要令牌)
//	GET  /api/v1/products              商品列表
//	GET  /api/v1/products/:id          商品详情
//
// **每加一条免鉴权路由都要问一次"它凭什么免鉴权"** ——
// 这个清单是安全边界,不是便利清单。

// PublicMiddleware 免鉴权组的占位中间件。
type PublicMiddleware struct{}

func NewPublicMiddleware() *PublicMiddleware {
	return &PublicMiddleware{}
}

func (m *PublicMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 刻意什么都不做:免鉴权是显式声明,不是忘了挂
		next(w, r)
	}
}
