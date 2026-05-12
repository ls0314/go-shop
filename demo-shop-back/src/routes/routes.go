package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

var userCtrl *handler.UserHandler

func InitUserModule() {
	userCtrl = handler.NewUserHandler()
}

func InitRoutes() *gin.Engine {
	r := gin.Default()
	// ========== 用户模块初始化 ==========
	InitUserModule()
	// ========== 用户信息模块初始化 ==========
	InitUserInfoModule()
	// ========== 注册用户信息路由 ==========
	RegisterUserInfoRoutes(r)
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
	// ========== 部门模块初始化 ==========
	InitDeptModule()
	// ========== 注册部门路由 ==========
	RegisterDeptRoutes(r)
	// ========== 数据权限模块初始化 ==========
	InitScopeModule()
	// ========== 注册数据权限路由 ==========
	RegisterScopeRoutes(r)
	// ========== 角色权限关联模块初始化 ==========
	InitRolePermModule()
	// ========== 注册角色权限关联路由 ==========
	RegisterRolePermRoutes(r)
	// ========== 用户角色关联模块初始化 ==========
	InitUserRoleModule()
	// ========== 注册用户角色关联路由 ==========
	RegisterUserRoleRoutes(r)
	// ========== 角色菜单关联模块初始化 ==========
	InitRoleMenuModule()
	// ========== 注册角色菜单关联路由 ==========
	RegisterRoleMenuRoutes(r)
	// ========== 用户部门关联模块初始化 ==========
	InitUserDeptModule()
	// ========== 注册用户部门关联路由 ==========
	RegisterUserDeptRoutes(r)
	// ========== 菜单权限关联模块初始化 ==========
	InitMenuPermModule()
	// ========== 注册菜单权限关联路由 ==========
	RegisterMenuPermRoutes(r)

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
		public.POST("/register", userCtrl.RegisterHandler)
		public.POST("/login", userCtrl.LoginHandler)
		public.POST("/refresh", userCtrl.RefreshHandler)
	}

	private := r.Group("/api/v1/user")
	private.Use(middleware.AuthMiddleware())
	{
		private.GET("/info", userCtrl.GetUserInfo)
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	return r
}
