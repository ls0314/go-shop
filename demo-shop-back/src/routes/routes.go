package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()
	// ========== 权限模块初始化 ==========
	InitPermissionModule()
	// ========== 注册权限路由 ==========
	RegisterPermissionRoutes(r)
	// ========== 菜单模块初始化 ==========
	InitMenuModule()
	// ========== 注册菜单路由 ==========
	RegisterMenuRoutes(r)
	// ========== 角色模块初始化 ==========
	InitRoleModule()
	// ========== 注册角色路由 ==========
	RegisterRoleRoutes(r)
	// ========== 角色模块初始化 ==========
	InitDeptModule()
	// ========== 注册角色路由 ==========
	RegisterDeptRoutes(r)

	// 配置CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	public := r.Group("/api/v1/user")
	{
		public.POST("/register", handler.RegisterHandler)
		public.POST("/login", handler.LoginHandler)
		public.POST("/refresh", handler.RefreshHandler)
	}

	private := r.Group("/api/v1/user")
	private.Use(middleware.AuthMiddleware())
	{
		private.GET("/info", handler.GetUserInfo)
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	return r
}
