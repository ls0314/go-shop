package main

import (
	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/config"
	"demo-shop/services/trade/internal/infra/lock"
	"demo-shop/services/trade/internal/infra/mq"
	cartserviceserver "demo-shop/services/trade/internal/server/cartservice"
	orderserviceserver "demo-shop/services/trade/internal/server/orderservice"
	paymentserviceserver "demo-shop/services/trade/internal/server/paymentservice"
	"demo-shop/services/trade/internal/svc"
	"demo-shop/services/trade/internal/task"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/trade.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// serviceGroup 同时托管 RPC 与后台任务,与 product-service 同构。
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1_tradev1.RegisterCartServiceServer(grpcServer, cartserviceserver.NewCartServiceServer(ctx))
		v1_tradev1.RegisterOrderServiceServer(grpcServer, orderserviceserver.NewOrderServiceServer(ctx))
		v1_tradev1.RegisterPaymentServiceServer(grpcServer, paymentserviceserver.NewPaymentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	}))

	addBackgroundTasks(group, c, ctx)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}

// addBackgroundTasks 按配置装配后台任务
func addBackgroundTasks(group *service.ServiceGroup, c config.Config, ctx *svc.ServiceContext) {
	// ---- 1. 延迟取消链路:outbox 投递器(发)+ 死信队列消费端(收)----
	//
	// 两套独立机制,都在时才是真正的"双保险":
	//
	//	投递器 + 消费端:消息在 expire_at 时刻转入死信队列 → 立即取消(精确到秒)
	//	扫描器:        每分钟扫一次过期未支付的订单     → 本地取消(粗粒度)
	//
	// 前一路失效时功能不中断,只是取消精度退化到扫描周期(1 分钟)。
	if c.Task.OutboxInterval > 0 {
		if ctx.MQ == nil {
			logx.Error("延迟取消链路未装配:RabbitMQ 未配置。" +
				"outbox 投递器与消费端都不启动,订单超时取消仅由定时扫描兜底")
		} else {
			// 发:投递器用自己的发布 channel
			dispatcher := task.NewOutboxDispatcher(ctx.OutboxRepo, ctx.MQ,
				lock.NewTaskLockManager(ctx.Redis), c.Task.OutboxInterval)
			group.Add(task.NewLoopService("outbox-dispatcher", dispatcher.Run))

			// 收:消费端另开一条 channel。
			// 开不起来不是致命错误(扫描仍能收敛),故只告警不阻断启动
			consumer, err := ctx.MQ.NewOrderDelayConsumer(mq.QueueOrderTradeDead, 1)
			if err != nil {
				logx.Errorf("开启延迟取消消费端失败(超时取消仍由定时扫描兜底): %v", err)
			} else {
				group.Add(task.NewLoopService("order-delay-consumer",
					task.NewOrderDelayConsumerService(consumer, ctx.OrderService).Run))
			}
		}
	} else {
		logx.Infof("延迟取消链路已按配置关闭(OutboxInterval=%v) —— 生产环境不应关闭", c.Task.OutboxInterval)
	}

	// ---- 2. 超时扫描(兜底那一路)----
	if c.Task.OrderScanInterval <= 0 {
		logx.Infof("订单超时扫描已按配置关闭(interval=%v) —— 生产环境不应关闭", c.Task.OrderScanInterval)
		return
	}
	group.Add(task.NewOrderTimeoutService(ctx.OrderRepo, ctx.OrderService, ctx.Redis).
		AsService(c.Task.OrderScanInterval))
}
