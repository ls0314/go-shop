package main

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/config"
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

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
