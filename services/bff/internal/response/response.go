// Package response 对外 HTTP 响应信封。

package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Envelope 对外响应信封。
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// SuccessMessage 成功时的固定文案(前端按它判成功,不可改)
const SuccessMessage = "Success"

// RateLimitedMessage 限流拒绝文案(与单体 middleware/reject 逐字一致)
const RateLimitedMessage = "请求过于频繁,请稍后重试"

// OK 成功响应(200)
func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Envelope{
		Code:    http.StatusOK,
		Message: SuccessMessage,
		Data:    data,
	})
}

// BadRequest 业务/参数失败(400)
func BadRequest(w http.ResponseWriter, code int, message string) {
	write(w, http.StatusBadRequest, Envelope{Code: code, Message: message})
}

// InternalError 服务端错误(500)
func InternalError(w http.ResponseWriter, code int, message string) {
	write(w, http.StatusInternalServerError, Envelope{Code: code, Message: message})
}

// Unavailable 下游不可用(503)
//
// 契约:message 以 " <服务名> 不可用" 结尾 —— 单体的接线冒烟测试按这个
// 后缀判"接线断了"而不是"依赖没起来"。改文案要同步改那里的判据。
func Unavailable(w http.ResponseWriter, message string) {
	write(w, http.StatusServiceUnavailable, Envelope{
		Code:    http.StatusServiceUnavailable,
		Message: message,
	})
}

// TooManyRequests 限流拒绝(429)
//
// 带 Retry-After:限流的礼貌语义是告诉客户端"什么时候可以再来"。
// 给保守的 1 秒 —— 精确值需算下一个令牌到期时刻,业务上秒级粒度足够。
func TooManyRequests(w http.ResponseWriter, message string) {
	w.Header().Set("Retry-After", "1")
	write(w, http.StatusTooManyRequests, Envelope{
		Code:    http.StatusTooManyRequests,
		Message: message,
	})
}

// write 统一写出。
//
// 编码失败只记状态码:此刻响应头已发出,再改状态码会 panic
// (http.ResponseWriter 的经典坑),只能让它以一个不完整的 body 结束。
func write(w http.ResponseWriter, status int, body Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// ErrDownstreamUnavailable 下游服务不可用的哨兵错误。
var ErrDownstreamUnavailable = errors.New("下游服务不可用")

// ErrUnauthorized 凭据无效 / 未登录。
var ErrUnauthorized = errors.New("未授权")

// unavailableSuffix 下游不可用文案的后缀(跨服务契约)。
const unavailableSuffix = " 不可用"

// InvalidParamMessage 参数错误的对外文案。
const InvalidParamMessage = "请求参数错误"

// Failure 错误出口:决定状态码与响应体。

// 下游不可用 → 503(接线断了,不是我错了)
// 凭据无效   → 401(前端跳登录页)
// 业务失败   → 400(与单体一致,前端按状态码分流)
// 其它       → 500(本服务的 bug 或未预期的故障,运维要看)
func Failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDownstreamUnavailable), hasUnavailableSuffix(err):
		Unavailable(w, err.Error())

	case errors.Is(err, ErrUnauthorized):
		Unauthorized(w, err.Error())

	case isBizError(err):
		// 与单体一致回 400。code 也用 400 —— 单体 utils.Fail(c, 400, msg)
		// 的 code 就是 400。
		BadRequest(w, http.StatusBadRequest, err.Error())

	default:
		// 保留原始文案:单体 utils.Error(c, 500, err.Error()) 就是这么做的,
		// 前端有些页面会展示它。改成通用文案会丢信息。
		InternalError(w, http.StatusInternalServerError, err.Error())
	}
}

// Unauthorized 凭据无效(401)
func Unauthorized(w http.ResponseWriter, message string) {
	write(w, http.StatusUnauthorized, Envelope{
		Code:    http.StatusUnauthorized,
		Message: message,
	})
}

// ErrInvalidParam 请求参数不合法(绑定或校验失败)。

var ErrInvalidParam = errors.New("请求参数错误")

// bizErrors 可预期的业务失败白名单。
var bizErrors = []error{
	ErrInvalidParam,
}

func isBizError(err error) bool {
	if err == nil {
		return false
	}
	for _, target := range bizErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// hasUnavailableSuffix 文案兜底(gRPC 自身的 Unavailable 不是我们的哨兵)
func hasUnavailableSuffix(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return len(msg) >= len(unavailableSuffix) &&
		msg[len(msg)-len(unavailableSuffix):] == unavailableSuffix
}
