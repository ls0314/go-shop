// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type DownstreamConfig struct {
	EtcdKey string `json:",default=user-service"`
}
type Config struct {
	rest.RestConf

	// JwtPublicKeyPath JWT 验签公钥(PEM 文件路径
	JwtPublicKeyPath string

	// Etcd 服务发现
	Etcd EtcdConfig

	// 四个下游服务
	User      DownstreamConfig `json:",optional"`
	Product   DownstreamConfig `json:",optional"`
	Marketing DownstreamConfig `json:",optional"`
	Trade     DownstreamConfig `json:",optional"`

	RateLimit      RateLimitConfig
	TrustedProxies []string `json:",optional"`

	// AllowedOrigins 允许跨域的前端来源(CORS)。
	//
	// ============================================================
	// 为什么必须显式配 —— 漏了它的症状很像"接口没路由到"
	// ============================================================
	//
	// 前端跑在 http://localhost:5173(Vite),请求 BFF 的 :9003
	// 是**跨域**。浏览器要求响应带 Access-Control-Allow-Origin,
	// 否则**即使是 200 也会被浏览器丢弃** —— 前端只看到网络错误,
	// 与"这个接口不存在"的表现完全一样。
	//
	// 单体自带这个中间件(demo-shop-back/src/routes/routes.go 里
	// 的内联闭包,硬编码 http://localhost:5173),故从单体切到 BFF 时
	// 若忘了配 CORS,登录等所有前端调用都会失败,而 curl 直接打
	// BFF 却是通的 —— 这个对比会让排查方向偏掉。
	//
	// ============================================================
	// 与单体的两处差别(都是刻意的)
	// ============================================================
	//
	//	① 单体硬编码,这里走配置 —— 换前端端口/域名不用改代码
	//	② 单体只允许一个来源(Header().Set 覆盖),这里支持多个
	//	   —— 本地 + 测试环境可以同时配
	//
	// 另外 go-zero 的 CORS 实现会自动处理 OPTIONS 预检并返回 204,
	// 与单体的 `if Method == OPTIONS { AbortWithStatus(204) }` 等价。
	//
	// **生产环境要改成真实域名**。留 localhost 在生产等于不设防
	// (任何从 localhost 发起的页面都能带凭据打这个 API)。
	AllowedOrigins []string `json:",default=[\"http://localhost:5173\"]"`
}

type RateLimitConfig struct {
	// IPRPS 每个 IP 每秒补充的令牌数(默认 5)
	IPRPS int `json:",default=5"`
	// IPBurst 每个 IP 的桶容量 = 允许的突发上限(默认 10)
	IPBurst int `json:",default=10"`
}

type EtcdConfig struct {
	Hosts []string `json:",default=[\"127.0.0.1:2379\"]"`
}
