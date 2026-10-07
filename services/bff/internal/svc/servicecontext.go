package svc

import (
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	v1_productv1 "demo-shop/api/gen/product/v1"
	v1_tradev1 "demo-shop/api/gen/trade/v1"
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
	RequestMeta     rest.Middleware
	Auth            rest.Middleware
	PublicRateLimit rest.Middleware
	Public          rest.Middleware

	// ---- 下游 gRPC 客户端 ----
	//
	// **直接持有生成的客户端,不包 façade**:转的目标是 .api 生成的
	// types.* 而不是单体的 model.*,包一层等于为不同目标复刻同一批方法。

	// UserRPC / RBACRPC / AddressRPC 三个 gRPC service 由**同一个
	// user-service 进程**提供、同一个 etcd key 发现,故共用一条连接。
	UserRPC    v1_userv1.UserServiceClient
	RBACRPC    v1_userv1.RBACServiceClient
	AddressRPC v1_userv1.AddressServiceClient

	// CartRPC / OrderRPC / PaymentRPC 同理,三个都由 trade-service 提供。
	CartRPC    v1_tradev1.CartServiceClient
	OrderRPC   v1_tradev1.OrderServiceClient
	PaymentRPC v1_tradev1.PaymentServiceClient

	// CategoryRPC / ProductRPC / InventoryRPC 三个都由 product-service 提供。
	CategoryRPC  v1_productv1.CategoryServiceClient
	ProductRPC   v1_productv1.ProductServiceClient
	InventoryRPC v1_productv1.InventoryServiceClient

	// CouponRPC 由 marketing-service 提供。
	CouponRPC v1_marketingv1.CouponServiceClient
}

// NewServiceContext verifier 由 main 在启动时加载 —— 验签器需要它,
// 而加载失败应当在启动时 panic(验签不可用等于全站 401)。
//
// **注意签名与 goctl 生成的不同**(生成的是单参 c):
// 这是刻意的,见 .goctl/api/main.tpl。让 svc 自己加载公钥会把
// "路径配错"推迟成运行时某个请求的 401,而不是启动失败。
func NewServiceContext(c config.Config, verifier *auth.Verifier) *ServiceContext {
	// 两个下游各建一条连接 —— 服务发现按 etcd key,一个 key 一条连接。
	//
	// 注意 **user-service 与 trade-service 是两条连接**:它们的 etcd key
	// 不同(user-service / trade-service),即使将来部署在同一台机器上
	// 也应当分开建 —— 共用一条会让"其中一个实例下线"影响另一个的
	// 负载均衡。
	userConn := rpc.Connect(c.Etcd.Hosts, c.User.EtcdKey)
	tradeConn := rpc.Connect(c.Etcd.Hosts, c.Trade.EtcdKey)
	productConn := rpc.Connect(c.Etcd.Hosts, c.Product.EtcdKey)
	marketingConn := rpc.Connect(c.Etcd.Hosts, c.Marketing.EtcdKey)

	return &ServiceContext{
		Config:          c,
		RequestMeta:     middleware.NewRequestMetaMiddleware(&c).Handle,
		Auth:            middleware.NewAuthMiddleware(verifier).Handle,
		PublicRateLimit: middleware.NewPublicRateLimitMiddleware(&c).Handle,
		Public:          middleware.NewPublicMiddleware().Handle,

		UserRPC:    v1_userv1.NewUserServiceClient(userConn),
		RBACRPC:    v1_userv1.NewRBACServiceClient(userConn),
		AddressRPC: v1_userv1.NewAddressServiceClient(userConn),

		CartRPC:    v1_tradev1.NewCartServiceClient(tradeConn),
		OrderRPC:   v1_tradev1.NewOrderServiceClient(tradeConn),
		PaymentRPC: v1_tradev1.NewPaymentServiceClient(tradeConn),

		CategoryRPC:  v1_productv1.NewCategoryServiceClient(productConn),
		ProductRPC:   v1_productv1.NewProductServiceClient(productConn),
		InventoryRPC: v1_productv1.NewInventoryServiceClient(productConn),

		CouponRPC: v1_marketingv1.NewCouponServiceClient(marketingConn),
	}
}
