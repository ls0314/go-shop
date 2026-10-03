package main

import (
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/config"
	couponserviceserver "demo-shop/services/marketing/internal/server/couponservice"
	"demo-shop/services/marketing/internal/svc"
	"demo-shop/services/marketing/internal/task"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/marketing.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// serviceGroup 同时托管 RPC 服务与后台任务:券闸门对账与表的所有权同进程
	// (DS-A-26 §3)。ServiceGroup 支持优雅退出,任务会结束当前轮次再退出。
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1_marketingv1.RegisterCouponServiceServer(grpcServer, couponserviceserver.NewCouponServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	}))

	addBackgroundTasks(group, c, ctx)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}

// addBackgroundTasks 按配置装配后台任务;周期 <= 0 表示不启动。
//
// Redis 不判空:NewServiceContext 用的是 redis.MustNewRedis,配置缺失时直接 panic,
// 到这里 Redis 必然非 nil(与 product-service 同一口径)。
func addBackgroundTasks(group *service.ServiceGroup, c config.Config, ctx *svc.ServiceContext) {
	if c.Tasks.CouponInterval <= 0 {
		logx.Infof("券闸门对账任务已按配置关闭(interval=%v)", c.Tasks.CouponInterval)
		return
	}
	group.Add(task.NewCouponReconcileService(ctx.CouponRepo, ctx.UserCouponRepo, ctx.Redis).
		AsService(c.Tasks.CouponInterval))
}
