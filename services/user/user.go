package main

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/config"
	addressserviceserver "demo-shop/services/user/internal/server/addressservice"
	rbacserviceserver "demo-shop/services/user/internal/server/rbacservice"
	userserviceserver "demo-shop/services/user/internal/server/userservice"
	"demo-shop/services/user/internal/svc"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1_userv1.RegisterRBACServiceServer(grpcServer, rbacserviceserver.NewRBACServiceServer(ctx))
		v1_userv1.RegisterUserServiceServer(grpcServer, userserviceserver.NewUserServiceServer(ctx))
		// 地址域独立成一个 gRPC service 而不是塞进 UserService:
		// 它是"用户自助维护的资料",与身份/RBAC 的关注点不同,
		// 而调用方(trade 取快照、BFF 的地址簿)只需要这一个域的方法。
		v1_userv1.RegisterAddressServiceServer(grpcServer, addressserviceserver.NewAddressServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
