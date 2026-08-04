// Package infra 基础设施层——统一管理所有外部中间件的生命周期
//
// 设计原则：
//   - 全局单例模式：各模块通过 GetXxx() 获取实例，避免函数签名链式传播
//   - 弱依赖降级：中间件不可用时不影响核心业务，仅打印 WARN 日志
//   - 统一初始化：main.go 中 InitInfra() 一次性初始化所有组件
package infra

import (
	"demo-shop-back/src/config"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/infra/pay"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// GlobalInfra 全局基础设施实例（main.go 中 InitInfra 赋值）
var GlobalInfra *Infra

// Infra 基础设施聚合——持有所有中间件连接
type Infra struct {
	MQ    *mq.RabbitMQ // RabbitMQ 连接（nil 表示未初始化或不可用）
	Redis *cache.RedisService
}

// Config 基础设施初始化配置
type Config struct {
	RabbitMQ config.RabbitMQConfig
	Redis    config.RedisConfig
}

// InitInfra 初始化所有基础设施组件
// 接收值：cfg - 基础设施配置（RabbitMQ DSN 等）
// 返回值：error - 初始化失败时返回错误（调用方可降级处理）
func InitInfra(cfg Config) error {
	var err error
	GlobalInfra = &Infra{}
	pay.InitGateways()

	// RabbitMQ 可选：DSN 为空时跳过，非空时初始化并声明拓扑
	if cfg.RabbitMQ.DSN != "" {
		GlobalInfra.MQ, err = mq.NewRabbitMQ(cfg.RabbitMQ.DSN)
		if err != nil {
			return err
		}
		if err := GlobalInfra.MQ.InitOrderDelayTopology(); err != nil {
			return fmt.Errorf("初始化订单延迟队列拓扑失败: %w", err)
		}
	}
	if cfg.Redis.Addr != "" {
		client, err := cache.NewRedisService(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err != nil {
			return err // 弱依赖，可降级
		}
		GlobalInfra.Redis = client
	}

	return nil
}

// GetCache 获取缓存服务（全局单例）
func GetCache() *cache.RedisService {
	if GlobalInfra == nil {
		return nil
	}
	return GlobalInfra.Redis
}

// GetMQ 获取默认 RabbitMQ 实例
// 接收值：无
// 返回值：*mq.RabbitMQ - 实例指针，未初始化时返回 nil
func GetMQ() *mq.RabbitMQ {
	if GlobalInfra == nil {
		return nil
	}
	return GlobalInfra.MQ
}

// StartOrderConsumer 启动订单超时消费者（goroutine）
// 接收值：无——消费者回调由 mq.RegisterCanceller 在 NewOrderService 中注入
func StartOrderConsumer() {
	if GlobalInfra != nil && GlobalInfra.MQ != nil {
		if err := GlobalInfra.MQ.StartOrderConsumer(); err != nil {
			log.Printf("[WARN] 启动订单消费者失败: %v", err)
		}
	}
}

// Shutdown 优雅关闭所有基础设施连接
// 接收值：无
func Shutdown() {
	if GlobalInfra != nil {
		if GlobalInfra.MQ != nil {
			GlobalInfra.MQ.Close()
		}
		if GlobalInfra.Redis != nil {
			GlobalInfra.Redis.Close()
		}
	}
}
