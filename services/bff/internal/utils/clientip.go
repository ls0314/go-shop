package utils

import (
	"context"
	"net"
	"net/http"
	"strings"

	"demo-shop/services/bff/internal/config"
)

// IndexHeaderDevice 设备标识的来源头。
const IndexHeaderDevice = "User-Agent"

// ClientIP 取真实客户端 IP
func ClientIP(r *http.Request, cfg *config.Config) string {
	if r == nil {
		return ""
	}

	remoteIP := hostOnly(r.RemoteAddr)

	if cfg == nil || !isTrusted(remoteIP, cfg.TrustedProxies) {
		return remoteIP
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return remoteIP
	}
	if idx := strings.IndexByte(xff, ','); idx >= 0 {
		xff = xff[:idx]
	}
	if ip := hostOnly(strings.TrimSpace(xff)); ip != "" {
		return ip
	}
	return remoteIP
}

// Device 取设备标识(原样返回 User-Agent)
func Device(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get(IndexHeaderDevice)
}

type requestMetaKey int

const requestMetaCtxKey requestMetaKey = iota

// requestMeta 一次请求的元信息
type requestMeta struct {
	clientIP string
	device   string
}

// WithRequestMeta 把请求元信息放进 context。
//
// **由中间件调用**(见 middleware/requestmeta),且该中间件挂在
// **全部四个路由组**上 —— 包括不鉴权的 PublicRateLimit(登录/注册)。
// 登录接口正需要 client_ip,而它不经过 Auth 中间件,故不能把这件事
// 挂在 Auth 上。
func WithRequestMeta(ctx context.Context, clientIP, device string) context.Context {
	return context.WithValue(ctx, requestMetaCtxKey, requestMeta{
		clientIP: clientIP,
		device:   device,
	})
}

// ClientIPFrom 从 context 取客户端 IP。
//
// **取不到返回空串,不报错**:这两个值只用于服务端的登录日志,
// 缺了不影响登录本身能否成功。为了一个日志字段让登录失败是错的取舍。
//
// 但"取不到"仍然是一个应当被发现的接线问题 —— 故
// 见 middleware/requestmeta 的说明与 handler 的注释。
func ClientIPFrom(ctx context.Context) string {
	if m, ok := ctx.Value(requestMetaCtxKey).(requestMeta); ok {
		return m.clientIP
	}
	return ""
}

// DeviceFrom 从 context 取设备标识(User-Agent 原样)
func DeviceFrom(ctx context.Context) string {
	if m, ok := ctx.Value(requestMetaCtxKey).(requestMeta); ok {
		return m.device
	}
	return ""
}

// isTrusted 判断 ip 是否落在任一可信网段内
func isTrusted(ip string, cidrs []string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range cidrs {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil && ipNet.Contains(parsed) {
			return true
		}
	}
	return false
}

// hostOnly 去掉端口
func hostOnly(addr string) string {
	if addr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	// 没有端口(或格式不标准):原样返回。
	//
	// 这里**不**再去剥方括号 —— 裸 IPv6 如 "::1" 本来就该原样保留,
	// 多一步"看起来更干净"的处理反而会把 "[::1]" 这种半成品
	// 变成一个不合法也不是原值的字符串。
	return addr
}
