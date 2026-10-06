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
//
// 用自定义类型而不是字符串,避免与其它包的键撞车(与 middleware 包
// 里 ctxKey 的做法一致)。
type ctxKeyRouteTemplate struct{}

// RouteTemplate 把路由模板注入 context。
//
// 模板是**注册期字面量** —— 由生成脚本从 routes.go 的 Path 字段取,
// 写成实参,不做任何"从请求路径反推模板"的动作。
//
// 为什么不能反推:go-zero 的 pathvar 只给参数**值**,分不清"字面量段"
// 与"参数"。实测 /api/v1/admin/menu/tree/tree 会被反推成
// /api/v1/admin/menu/:id/:id,而正确模板是 /:id/tree —— 反推在参数值
// 撞上字面量时必然出错。
//
// 注入模板的用途是日志/审计/按路由维度的指标 —— 判权不依赖它
// (权限码是编译期常量,见 Permission)。
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
//
// codes 由生成脚本在生成期从 user-service 查到(按 api_path +
// request_method),作为**字面量**写进 routes.go。所以运行时不需要
// "这个接口要什么权限码"的反查 —— 那条链被彻底消灭了。
//
// 语义是"任一满足"而不是"全部满足":与单体 PermissionMiddleware 一致
// (它遍历接口所需的所有码,命中一个即放行)。这也让"读权限与写权限
// 共用一个模板"这类配置成为可能。
//
// 空 codes 一律拒绝(fail-closed):生成脚本已经在生成期拦住了这种情况
// (查不到码就报错退出),走到这里说明代码被手工改过,此时拒绝比放行安全。
//
// 失败语义:
//
//	取不到身份   → 401(走到这里说明该路由没挂 Auth,是配置错误)
//	下游不可用   → 503(可重试),不会误报成"没权限"
//	无任一权限   → 403 + "用户无操作权限"
//
// 403 是对单体的**有意偏离**:单体 utils.Fail(c, 400, model.UserHasNotPerm)
// 给的是 400。403 语义更准,前端能据此区分"无权访问"与"参数错误";
// 若前端严格按 400 分支处理,把 Forbidden 换成 BadRequest 即可(一行)。
func Permission(codes []string, rbac v1_userv1.RBACServiceClient, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {
			// 未登录在 Auth 那层就该被拦下;走到这里说明该路由的中间件
			// 配置漏了 Auth —— 是配置错误,不是用户问题。
			logx.WithContext(r.Context()).Errorf(
				"判权取不到身份(该路由未挂 Auth?): %s %s", r.Method, RouteTemplateFrom(r.Context()))
			response.Unauthorized(w, "未登录")
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
