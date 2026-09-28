package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/es"
	"demo-shop-back/src/infra/mq"
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
