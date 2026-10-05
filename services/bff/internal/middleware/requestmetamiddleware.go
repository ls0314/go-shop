package middleware

import (
	"net/http"

	"demo-shop/services/bff/internal/config"
	"demo-shop/services/bff/internal/utils"
)

// RequestMetaMiddleware 把本次请求的 client_ip 与 device 放进 context。
//
// ============================================================
// 为什么它必须独立于 Auth,并且挂在**全部四个组**上
// ============================================================
//
// user-service 的 LoginReq 要 client_ip 与 device 用于写登录日志。
// 而登录接口挂在 **PublicRateLimit** 组上 —— **不经过 Auth 中间件**。
//
// 所以"把请求信息塞进 context"这件事不能寄生在 Auth 里:
// 那样登录拿不到 IP,而这恰恰是最需要记 IP 的接口(要能看出
// 某个 IP 在爆破)。
//
// 故本中间件与鉴权无关、对所有路由生效,且**必须排在 Auth 之前**
// (在 @server 的 middleware 列表里写在前面的先执行),
// 这样 Auth 拒绝请求时日志里也有 IP 可记。
//
// ============================================================
// 为什么不做成"全局中间件"
// ============================================================
//
// go-zero 的 .api 没有引擎级中间件 —— 所有中间件都挂在 @server 组上。
// 因此本中间件出现在全部四个 @server 块里:
//
//	Auth / Public / PublicRateLimit / AuthRateLimit
//
// 每加一个新的中间件组,都要记得带上它。**漏掉的后果是那个组的
// 登录日志字段为空** —— 而客户端与前端都察觉不到,只有翻日志时
// 才会发现。这是本项目里最容易被漏的一处接线。
type RequestMetaMiddleware struct {
	cfg *config.Config
}

// NewRequestMetaMiddleware cfg 用于可信代理判定(决定是否采信
// X-Forwarded-For)。传 nil 时按"不信任任何代理"处理(安全默认)。
func NewRequestMetaMiddleware(cfg *config.Config) *RequestMetaMiddleware {
	return &RequestMetaMiddleware{cfg: cfg}
}

func (m *RequestMetaMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取规则与按 IP 限流**共用** utils.ClientIP ——
		// 两处判"客户端是谁"必须一致,否则会出现
		// "限流按 A 判、日志记 B"这种排查时会怀疑人生的现象。
		ctx := utils.WithRequestMeta(
			r.Context(),
			utils.ClientIP(r, m.cfg),
			utils.Device(r),
		)
		next(w, r.WithContext(ctx))
	}
}
