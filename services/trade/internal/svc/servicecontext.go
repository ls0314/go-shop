package svc

import (
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/trade/internal/config"
	"demo-shop/services/trade/internal/infra/couponclient"
	"demo-shop/services/trade/internal/infra/mq"
	"demo-shop/services/trade/internal/infra/productclient"
	"demo-shop/services/trade/internal/repository"
	"demo-shop/services/trade/internal/service"
	"demo-shop/services/trade/internal/utils"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ServiceContext 依赖注入容器
type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Redis

	CartRepo    *repository.CartItemRepo
	OrderRepo   *repository.OrderRepo
	PaymentRepo *repository.PaymentRepo
	OutboxRepo  *repository.OutboxRepo

	// 下游客户端。一个 productclient 同时实现 InventoryRPC 与 ProductReadRPC ——
	// 两者是同一个服务、同一个 etcd key、同一条连接(见 infra/productclient 的说明)。
	InventoryRPC service.InventoryRPC
	ProductRPC   service.ProductReadRPC
	CouponRPC    service.CouponRPC

	// MQ 订单域的 RabbitMQ 客户端(投递延迟取消消息)。
	MQ *mq.Client

	OrderService   *service.OrderService
	CartService    *service.CartService
	PaymentService *service.PaymentService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接 trade_db 失败：" + err.Error())
	}

	// product-service:一条连接、两个域的客户端
	inventoryConn, productConn := newProductConn(c)
	prodClient := productclient.NewClient(inventoryConn, productConn)

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  redis.MustNewRedis(c.Redis.RedisConf),

		CartRepo:    repository.NewCartItemRepo(db),
		OrderRepo:   repository.NewOrderRepo(db),
		PaymentRepo: repository.NewPaymentRepo(db),
		OutboxRepo:  repository.NewOutboxRepo(db),

		InventoryRPC: prodClient,
		ProductRPC:   prodClient,
		CouponRPC:    couponclient.NewClient(newMarketingConn(c)),
	}

	sf := initSnowflake(c)

	ctx.OrderService = service.NewOrderService(ctx.OrderRepo, ctx.CartRepo, ctx.OutboxRepo,
		ctx.InventoryRPC, ctx.CouponRPC, ctx.ProductRPC, sf)
	ctx.CartService = service.NewCartService(ctx.CartRepo, ctx.ProductRPC)
	ctx.PaymentService = service.NewPaymentService(ctx.OrderRepo, ctx.PaymentRepo,
		ctx.InventoryRPC, ctx.OrderService, sf)

	return ctx
}

// newMQClient 建连 RabbitMQ 并声明订单域拓扑。
func newMQClient(c config.Config) *mq.Client {
	if c.RabbitMQ.Host == "" {
		logx.Info("未配置 RabbitMQ,outbox 投递器不启动(超时取消由定时扫描兜底)")
		return nil
	}
	client, err := mq.NewClient(c.RabbitMQ.URL())
	if err != nil {
		logx.Errorf("连接 RabbitMQ 失败,outbox 投递器不启动(超时取消仅由定时扫描兜底): %v", err)
		return nil
	}
	return client
}

// initSnowflake 初始化订单号生成器。
func initSnowflake(c config.Config) *utils.Snowflake {
	workerId := c.WorkerId
	if workerId <= 0 {
		workerId = 1
	}
	sf, err := utils.NewSnowflake(workerId)
	if err != nil {
		// 配置错误直接 panic:带错 workerId 跑起来会生成重复订单号,
		// 那比启动失败难查得多(fail-fast)
		panic("初始化雪花生成器失败：" + err.Error())
	}
	return sf
}

// newProductConn 建连 product-service,返回两个域的客户端(同一条连接)。
func newProductConn(c config.Config) (v1_productv1.InventoryServiceClient, v1_productv1.ProductServiceClient) {
	key := c.Product.EtcdKey
	if key == "" {
		key = "product-service"
	}
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: key},
	})
	if err != nil {
		logx.Errorf("连接 product-service 失败,下单锁库存与购物车回填将不可用: %v", err)
		return nil, nil
	}
	conn := client.Conn()
	return v1_productv1.NewInventoryServiceClient(conn), v1_productv1.NewProductServiceClient(conn)
}

// newMarketingConn 建连 marketing-service
func newMarketingConn(c config.Config) v1_marketingv1.CouponServiceClient {
	key := c.Marketing.EtcdKey
	if key == "" {
		key = "marketing-service"
	}
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: key},
	})
	if err != nil {
		logx.Errorf("连接 marketing-service 失败,下单核销券与取消退券将不可用: %v", err)
		return nil
	}
	return v1_marketingv1.NewCouponServiceClient(client.Conn())
}
