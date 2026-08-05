package main

import (
	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/routes"
	"demo-shop-back/src/utils"
	"log"
)

func main() {
	// 加载配置文件
	err := config.LoadConfig("resource/application.yaml")
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

	// 雪花算法初始化
	utils.InitSnowflake(1)

	// 基础组件初始化（redis，mq等）
	if err := infra.InitInfra(infra.Config{
		RabbitMQ: config.GlobalConfig.RabbitMQ,
		Redis:    config.GlobalConfig.Redis,
	}); err != nil {
		log.Printf("[WARN] 基础设施初始化失败: %v", err)
	}
	defer infra.Shutdown()

	// 操作日志初始化
	middleware.InitLogWorker()
	// JWT初始化
	middleware.InitJWT(utils.Secret)
	// 路由初始化
	router := routes.InitRoutes()

	// 启动mq消费者
	infra.StartOrderConsumer()

	err = router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return
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
