package main

import (
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/config"
	categoryserviceserver "demo-shop/services/product/internal/server/categoryservice"
	inventoryserviceserver "demo-shop/services/product/internal/server/inventoryservice"
	productserviceserver "demo-shop/services/product/internal/server/productservice"
	"demo-shop/services/product/internal/svc"
	"demo-shop/services/product/internal/task"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/product.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// serviceGroup 同时托管 RPC 服务与后台任务:
	// 拆库后 ES 索引同步与两个对账任务归 product-service(DS-A-26 §3),
	// 它们此前跑在单体里读 demo_shop —— 现在与表的所有权在同一个进程内。
	//
	// 为什么不用裸 go 起 goroutine:ServiceGroup 支持优雅退出,
	// 任务收到取消信号会结束当前轮次再退出,不会被滚动发布硬砍在半途。
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1_productv1.RegisterInventoryServiceServer(grpcServer, inventoryserviceserver.NewInventoryServiceServer(ctx))
		v1_productv1.RegisterCategoryServiceServer(grpcServer, categoryserviceserver.NewCategoryServiceServer(ctx))
		v1_productv1.RegisterProductServiceServer(grpcServer, productserviceserver.NewProductServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	}))

	addBackgroundTasks(group, c, ctx)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}

// addBackgroundTasks 按配置装配后台任务;周期 <= 0 表示不启动该任务。
//
// 关于 ctx.Redis 为什么不用判空:NewServiceContext 里用的是 redis.MustNewRedis,
// 配置缺失时它直接 panic,所以到这里 Redis 必然非 nil。
// ES 则确实可能为 nil —— 它是可选依赖(Addresses 留空 = 不启用)。
func addBackgroundTasks(group *service.ServiceGroup, c config.Config, ctx *svc.ServiceContext) {
	// ---- ES 对账 ----
	// 没配 ES = 没有索引可对账,跳过;这与"配了地址却连不上"不同,
	// 后者在 NewServiceContext 里已 fail-fast。
	switch {
	case c.Task.ESIncInterval <= 0 || c.Task.ESFullInterval <= 0:
		logx.Infof("ES 对账任务已按配置关闭(inc=%v full=%v)", c.Task.ESIncInterval, c.Task.ESFullInterval)
	case ctx.ES == nil:
		logx.Info("ES 未配置,不启动 ES 对账任务")
	default:
		group.Add(task.NewReconcileService(ctx.ProductRepo, ctx.ES, ctx.Redis).
			AsService(c.Task.ESIncInterval, c.Task.ESFullInterval))
	}

	// ---- 库存闸门对账 ----
	if c.Task.StockInterval <= 0 {
		logx.Infof("库存闸门对账任务已按配置关闭(interval=%v)", c.Task.StockInterval)
		return
	}
	group.Add(task.NewStockReconcileService(ctx.ProductRepo, ctx.Redis).
		AsService(c.Task.StockInterval))
}
