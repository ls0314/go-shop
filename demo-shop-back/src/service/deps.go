package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/es"
	"demo-shop-back/src/infra/mq"

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
}

// NewServiceDeps 在 composition root 读一次全局依赖。
func NewServiceDeps() ServiceDeps {
	return ServiceDeps{
		DB:        db.DB,
		Cache:     infra.GetCache(),
		GateCache: infra.GetGateCache(),
		MQ:        infra.GetMQ(),
		ES:        infra.GetES(),
	}
}
