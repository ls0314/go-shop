// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/pkg/auth"
	"demo-shop/services/bff/internal/config"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config config.Config
	// 中间件按路由组的鉴权要求绑定。
	//
	// 字段名**必须与 bff.api 里 middleware: 后的标识符逐字一致** ——
	// 生成物 routes.go 用 serverCtx.<名字> 引用它们,改名即编译失败。
	//
	// RequestMeta 排在最前:它把 client_ip / device 放进 context,
	// 且对**所有**路由生效(包括不鉴权的登录/注册)。
	// 见 middleware/requestmetamiddleware.go 的说明。
	RequestMeta     rest.Middleware
	Auth            rest.Middleware
	PublicRateLimit rest.Middleware
	Public          rest.Middleware

	// ---- 下游 gRPC 客户端 ----

	// UserRPC 用户域客户端
	UserRPC v1_userv1.UserServiceClient
	// RBACRPC RBAC域客户端
	RBACRPC v1_userv1.RBACServiceClient
	// AddressRPC 地址域客户端
	AddressRPC v1_userv1.AddressServiceClient
}

func NewServiceContext(c config.Config, verifier *auth.Verifier) *ServiceContext {

	// user-service服务连接
	userConn := rpc.Connect(c.Etcd.Hosts, c.User.EtcdKey)

	return &ServiceContext{
		Config:          c,
		RequestMeta:     middleware.NewRequestMetaMiddleware(&c).Handle,
		Auth:            middleware.NewAuthMiddleware(verifier).Handle,
		PublicRateLimit: middleware.NewPublicRateLimitMiddleware(&c).Handle,
		Public:          middleware.NewPublicMiddleware().Handle,

		UserRPC:    v1_userv1.NewUserServiceClient(userConn),
		RBACRPC:    v1_userv1.NewRBACServiceClient(userConn),
		AddressRPC: v1_userv1.NewAddressServiceClient(userConn),
	}
}
