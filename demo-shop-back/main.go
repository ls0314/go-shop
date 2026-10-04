package main

import (
	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/routes"
	"demo-shop-back/src/service"
	"demo-shop-back/src/task"
	"demo-shop-back/src/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "resource/application.yaml"
	}
	// 加载配置文件
	err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	log.Println("配置文件加载成功")

	// 执行数据库迁移
	err = db.RunMigrations()
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 程序结束时关闭数据库连接
	defer func() {
		// 关闭数据库连接
		if err := db.CloseDB(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		} else {
			log.Println("数据库连接已关闭")
		}
	}()

	// 种子数据:演示实例专用(版本账本机制,见 db/seed.go)。
	// compose 注入 DEMO_SHOP_SEED_DIR;本地 go run 不设该变量则不加载。
	// 每次启动都会检查账本,已应用的种子自动跳过,新增种子文件放 db/seeds/ 即可。
	if seedDir := os.Getenv("DEMO_SHOP_SEED_DIR"); seedDir != "" {
		if err := db.RunSeeds(seedDir, config.GlobalConfig.Database.GetDSN()); err != nil {
			log.Fatalf("种子数据加载失败: %v", err)
		}
		log.Println("种子数据检查完成:", seedDir)
	}

	// 雪花算法初始化
	workerId := int64(1)
	if v := os.Getenv("DEMO_SHOP_SNOWFLAKE_WORKER_ID"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			workerId = id
		}
	}
	utils.InitSnowflake(workerId)

	// 基础组件初始化（redis，mq等）
	if err := infra.InitInfra(infra.Config{
		RabbitMQ: config.GlobalConfig.RabbitMQ,
		Redis:    config.GlobalConfig.Redis,
		ES:       config.GlobalConfig.ES,
	}); err != nil {
		log.Printf("[WARN] 基础设施初始化失败: %v", err)
	}
	defer infra.Shutdown()

	// 服务层依赖集合:全局读取点在此收敛为一处(B0 去全局化)
	// 必须放在 infra.InitInfra 之后 —— NewServiceDeps 会读取 infra 单例与 db.DB
	deps := service.NewServiceDeps()

	// 对账任务初始化
	task.Init(deps)

	// 可观测性:独立内部端口暴露 /metrics + DB 连接池水位采样
	metrics.StartMetricsServer()
	metrics.StartDBPoolSampler(30 * time.Second)

	// 操作日志初始化
	middleware.InitLogWorker(repository.NewOperationLogRepo(deps.DB))

	// JWT 验签初始化:只加载公钥,私钥仅存于 user-service。
	// 默认路径按"工作目录 = demo-shop-back/"计算,退一级到仓库根。
	jwtPublicKeyPath := os.Getenv("DEMO_SHOP_JWT_PUBLIC_KEY_PATH")
	if jwtPublicKeyPath == "" {
		jwtPublicKeyPath = "../jwt_keys/dev_public.pem"
	}
	middleware.InitJWT(jwtPublicKeyPath)
	// 路由初始化
	router := routes.InitRoutes(deps)

	// **订单延迟消费者不再启动**。
	//
	// 它消费 order.delay.cancel 队列、回调本地 OrderService 做超时取消。
	// 订单三表已迁 trade_db,那个回调取消不到任何单 —— 更糟的是它会把
	// 消息从队列里**消费掉**:trade-service 的消费者(与 trade 侧 outbox
	// 投递器)就收不到了,等于超时取消被静默吞掉。
	//
	// 超时取消现在的归属:
	//   - trade 侧 outbox(trade_db.sys_outbox_message)写延迟消息 —— 投递器待实现;
	//   - trade 侧**超时扫描**(task/ordertimeout.go,每分钟一轮,带分布式锁)——
	//     已实现,是当前唯一生效的兜底。故投递器缺失期间功能不中断。
	//
	// 待 trade 侧的 outbox 投递器落地后,消费者也应在 trade-service 内启动。

	// main.go 原 SetTrustedProxies(["127.0.0.1"]) 处替换:
	trusted := os.Getenv("DEMO_SHOP_TRUSTED_PROXIES")
	if trusted == "" {
		trusted = "127.0.0.1" // 本地裸跑默认;compose 里注入 172.16.0.0/12
	}
	if err := router.SetTrustedProxies(strings.Split(trusted, ",")); err != nil {
		log.Printf("设置可信代理失败: %v", err)
	}

	serverPort := config.GlobalConfig.Server.Port
	if serverPort == "" {
		serverPort = "9000"
	}

	log.Printf("服务器启动在端口: %s", serverPort)
	err = router.Run(":" + serverPort)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
		return
	}
}
