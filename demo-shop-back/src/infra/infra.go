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
	"demo-shop-back/src/infra/es"
	"demo-shop-back/src/infra/mq"
	"demo-shop-back/src/infra/pay"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// GlobalInfra 全局基础设施实例（main.go 中 InitInfra 赋值）
var GlobalInfra *Infra

// Infra 基础设施聚合——持有所有中间件连接
type Infra struct {
	MQ    *mq.RabbitMQ // RabbitMQ 连接（nil 表示未初始化或不可用）
	Redis *cache.RedisService
	ES    *es.ESClient
}

// Config 基础设施初始化配置
type Config struct {
	RabbitMQ config.RabbitMQConfig
	Redis    config.RedisConfig
	ES       config.ESConfig
}

// InitInfra 初始化所有基础设施组件
// 接收值：cfg - 基础设施配置（RabbitMQ DSN 等）
// 返回值：error - 初始化失败时返回错误（调用方可降级处理）
func InitInfra(cfg Config) error {
	var err error
	GlobalInfra = &Infra{}
	pay.InitGateways()

	// RabbitMQ 可选：DSN 为空时跳过，非空时初始化。
	//
	// **不再声明订单延迟队列拓扑**。那套拓扑(order.dead./order.delay.queue)
	// 属于订单域,已随订单表迁到 trade-service —— trade 侧声明自己的
	// order.trade.* 拓扑(infra/mq/client.go),两边并存但互不干扰。
	//
	// 单体这里仍保留连接:它给自己的 sys_outbox_message 做投递
	// (见 task/init_recocile.go),那是**本库**的发件箱,与订单域无关。
	if cfg.RabbitMQ.DSN != "" {
		GlobalInfra.MQ, err = mq.NewRabbitMQ(cfg.RabbitMQ.DSN)
		if err != nil {
			log.Printf("[WARN] 启动mq失败: %v ,降级", err)
		} else {
			log.Printf("[INFO] MQ启动")
		}
	}

	if len(cfg.ES.Addresses) > 0 {
		client, err := es.NewESClient(cfg.ES.Addresses)
		if err != nil {
			log.Printf("[WARN] 启动es失败: %v ,降级", err)
		} else {
			GlobalInfra.ES = client
			log.Printf("[INFO] ES启动")
		}

	}

	if cfg.Redis.Addr != "" {
		client, err := cache.NewRedisService(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err != nil {
			log.Printf("[WARN] 启动redis失败: %v ,降级", err)
		} else {
			GlobalInfra.Redis = client
			log.Printf("[INFO] Redis启动")
		}
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

func GetES() *es.ESClient {
	if GlobalInfra == nil {
		return nil
	}
	return GlobalInfra.ES
}

// StartOrderConsumer 已删除。
//
// 它启动的消费者监听 order.dead.queue,回调做的是**本地**取消
// (service.OrderService.CancelOrderBySystem)—— 而订单三表已迁 trade_db,
// 本地那张表停更了。留着它有两个坏结果:
//
//  1. 取消不到任何单;
//  2. 更糟的是它会把消息 **Ack 掉**,于是 trade 侧的消费者永远收不到 ——
//     超时取消被静默吞掉,两边日志都正常。
//
// 超时取消的消费端现在在 trade-service(internal/task/orderdelayconsumer.go),
// 监听 trade 自己声明的 order.trade.dead.queue。

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

func GetGateCache() *cache.RedisService {
	if os.Getenv("DEMO_SHOP_GATE_ENABLED") == "false" {
		return nil
	}
	return GetCache()
}
