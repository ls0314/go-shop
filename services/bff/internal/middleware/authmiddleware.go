package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"demo-shop/services/bff/internal/auth"
	"demo-shop/services/bff/internal/response"
)

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyUsername
)

// UserID 从 context 取调用方用户 ID。
func UserID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(int64)
	return v, ok
}

// Username 从 context 取用户名(操作日志等场景要用)
func Username(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyUsername).(string)
	return v, ok
}

type AuthMiddleware struct {
	verifier *auth.Verifier
}

// NewAuthMiddleware verifier 由 main 在启动时加载(加载失败即 panic)。
func NewAuthMiddleware(verifier *auth.Verifier) *AuthMiddleware {
	return &AuthMiddleware{verifier: verifier}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		if raw == "" {
			response.BadRequest(w, http.StatusUnauthorized, "未登录")
			return
		}

		token := strings.TrimPrefix(raw, "Bearer ")
		if token == "" {
			response.BadRequest(w, http.StatusUnauthorized, "格式错误")
			return
		}

		claims, err := m.verifier.ParseAccessToken(token)
		if err != nil {
			response.BadRequest(w, http.StatusUnauthorized, authErrMessage(err))
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxKeyUsername, claims.Username)
		next(w, r.WithContext(ctx))
	}
}

// authErrMessage 把验签错误翻成前端提示。
func authErrMessage(err error) string {
	switch {
	case errors.Is(err, auth.ErrTokenExpired):
		return "token已过期，请重新登录"
	case errors.Is(err, auth.ErrTokenNotValidYet):
		return "token尚未生效"
	case errors.Is(err, auth.ErrTokenMalformed):
		return "token格式错误"
	default:
		return "无效的token"
	}
}
