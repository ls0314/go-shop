// Package guard 按路由挂的访问控制。
//
// 与 internal/middleware 的区别:中间件是**按路由组**挂的(在 bff.api 的
// @server 块里声明,由 svc 提供字段);而判权需要的权限码**同一组内也
// 不同**(订单组的"查列表"与"发货"是两个码),组级挂载表达不了,
// 所以它是**按路由**挂在处理器外层的。
//
// 两个函数都返回 http.HandlerFunc,正是 rest.Route.Handler 的类型,
// 所以能直接写在生成物的处理器表达式外层。
package guard

import (
	"context"
	"net/http"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/response"

	"github.com/zeromicro/go-zero/core/logx"
)

// ctxKeyRouteTemplate 路由模板在 context 里的键。
type ctxKeyRouteTemplate struct{}

// RouteTemplate 把路由模板注入 context。
func RouteTemplate(template string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyRouteTemplate{}, template)
		next(w, r.WithContext(ctx))
	}
}

// RouteTemplateFrom 从 context 取路由模板(取不到返回空串)。
func RouteTemplateFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRouteTemplate{}).(string)
	return v
}

// Permission 判权:调用方必须持有 codes 中的**任意一个**。
func Permission(codes []string, rbac v1_userv1.RBACServiceClient, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {

			logx.WithContext(r.Context()).Errorf(
				"判权取不到身份(该路由未挂 Auth?): %s %s", r.Method, RouteTemplateFrom(r.Context()))

			response.InternalError(w, http.StatusInternalServerError,
				"服务配置错误:该接口未挂认证中间件")
			return
		}

		if len(codes) == 0 {
			logx.WithContext(r.Context()).Errorf(
				"接口未配置权限码(fail-closed 拒绝): %s %s", r.Method, RouteTemplateFrom(r.Context()))
			response.Forbidden(w, response.PermissionNotConfiguredMessage)
			return
		}

		ctx, cancel := rpc.CtxWithTimeout(r.Context())
		defer cancel()

		resp, grpcErr := rbac.ListPermCodesByUserId(ctx, &v1_userv1.ListPermCodesByUserIdReq{
			UserId: userID,
		})

		// 这个 RPC **没有业务错误通道**(proto 注释:"无任何权限时返回空
		// 数组,不返回错误"),所以只有"通"与"下游挂了"两种结果,
		// errorMsg 传空串即可。
		if _, err := rpc.Classify("", grpcErr); err != nil {
			// 下游不可用 → 503。**不能当成"没权限"** —— 那会让用户以为
			// 自己被降权,而实际只是依赖抖了一下。
			response.Failure(w, err)
			return
		}

		if !hasAny(resp.GetPermCodes(), codes) {
			response.Forbidden(w, response.PermissionDeniedMessage)
			return
		}

		next(w, r)
	}
}

// hasAny 判断 owned 里是否至少有一个落在 required 里。
func hasAny(owned, required []string) bool {
	if len(owned) == 0 || len(required) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(owned))
	for _, c := range owned {
		set[c] = struct{}{}
	}
	for _, c := range required {
		if _, ok := set[c]; ok {
			return true
		}
	}
	return false
}
