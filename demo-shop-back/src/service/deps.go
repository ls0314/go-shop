package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/es"
	"demo-shop-back/src/infra/inventoryclient"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/infra/productclient"
	"demo-shop-back/src/infra/userclient"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

// ServiceDeps 服务层依赖集合(composition root 构造一次,注入所有 services)。
// 语义约定:
//   - DB 必填;其余可为 nil,表示该中间件未配置,调用方必须判空(降级路径);
//   - Cache 与 GateCache 是**两个不同语义**的字段,不可混用:
//     Cache 是通用缓存;GateCache 额外承载 DEMO_SHOP_GATE_ENABLED 熔断语义
type ServiceDeps struct {
	DB        *gorm.DB
	Cache     *cache.RedisService
	GateCache *cache.RedisService
	MQ        *mq.RabbitMQ
	ES        *es.ESClient
	UserRPC   *userclient.PermCodesClient
	// InventoryRPC 库存四操作。类型是接口而非具体客户端,以便测试注入记录桩
	// (见 inventory_rpc.go 的说明);生产实现仍是 *inventoryclient.InventoryClient。
	InventoryRPC InventoryStockRPC
	ProductRPC   *productclient.ProductClient
}

// NewServiceDeps 在 composition root 读一次全局依赖。
func NewServiceDeps() ServiceDeps {
	deps := ServiceDeps{
		DB:        db.DB,
		Cache:     infra.GetCache(),
		GateCache: infra.GetGateCache(),
		MQ:        infra.GetMQ(),
		ES:        infra.GetES(),
	}
	// user-service 走 etcd 服务发现;连不上不阻断启动,判权链路降级为报错。
	client, err := userclient.NewPermCodesClient(
		envList("DEMO_SHOP_ETCD_HOSTS", "127.0.0.1:2379"),
		getEnv("DEMO_SHOP_ETCD_KEY", "user-service"),
		deps.Cache,
	)
	if err != nil {
		log.Printf("[WARN] 连接 user-service 失败,判权将不可用: %v", err)
	} else {
		deps.UserRPC = client
	}

	// product-service 同理:库存四操作走 RPC,连不上不阻断启动,下单与支付会报错。
	invClient, err := inventoryclient.NewInventoryClient(
		envList("DEMO_SHOP_ETCD_HOSTS", "127.0.0.1:2379"),
		getEnv("DEMO_SHOP_PRODUCT_ETCD_KEY", "product-service"),
	)
	if err != nil {
		log.Printf("[WARN] 连接 product-service 失败,库存操作将不可用: %v", err)
	} else {
		deps.InventoryRPC = invClient
	}

	// 商品/类目读写同样走 RPC(与库存同一个服务,复用 etcd key)。
	prodClient, err := productclient.NewProductClient(
		envList("DEMO_SHOP_ETCD_HOSTS", "127.0.0.1:2379"),
		getEnv("DEMO_SHOP_PRODUCT_ETCD_KEY", "product-service"),
	)
	if err != nil {
		log.Printf("[WARN] 连接 product-service 失败,商品与类目接口将不可用: %v", err)
	} else {
		deps.ProductRPC = prodClient
	}
	return deps
}

// envList 读逗号分隔的列表型环境变量。
func envList(name, fallback string) []string {
	v := getEnv(name, fallback)
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// getEnv 读环境变量,空值时用默认值。
func getEnv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
