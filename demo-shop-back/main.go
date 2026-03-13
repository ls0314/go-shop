package main

import (
	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"log"

	"github.com/gin-gonic/gin"
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
	log.Println("数据库迁移成功")

	// 初始化数据库连接
	err = db.InitDB()
	if err != nil {
		log.Fatalf("初始化数据库连接失败: %v", err)
	}
	log.Println("数据库连接初始化成功")

	// 程序结束时关闭数据库连接
	defer func() {
		// 关闭数据库连接
		if err := db.CloseDB(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		} else {
			log.Println("数据库连接已关闭")
		}
	}()

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()
	api := r.Group("/api")
	api.POST("/register", handler.RegisterHandler)
	api.POST("/login", handler.LoginHandler)

	auth := api.Group("/user")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/info", handler.UserInfo)

	err = r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return
	}
	// 配置CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	serverPort := config.GlobalConfig.Server.Port
	if serverPort == "" {
		serverPort = "9001"
	}

	log.Printf("服务器启动在端口: %s", serverPort)
	err = r.Run(":" + serverPort)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
		return
	}
}
