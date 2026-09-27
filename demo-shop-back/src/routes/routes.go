package routes

import (
	"demo-shop-back/src/service"
	"net/http"

	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes(deps service.ServiceDeps) *gin.Engine {
	r := gin.Default()

	// HTTP 层指标:第一个业务中间件,保证所有请求(含404/预检)被计量
	r.Use(middleware.HTTPMetrics())

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

	// 全局兜底限流(DS-A-21):保护 DB/下游总容量;healthz 豁免
	r.Use(middleware.GlobalRateLimit())

	// ========== 用户模块初始化 ==========
	InitUserModule(deps)
	// ========== 注册用户路由 ==========
	RegisterUserRoutes(r)
	// ========== 用户信息模块初始化 ==========
	InitUserInfoModule(deps)
	// ========== 注册用户信息路由 ==========
	RegisterUserInfoRoutes(r)
	// ========== 权限模块初始化 ==========
	InitPermissionModule(deps)
	// ========== 注册权限路由 ==========
	RegisterPermissionRoutes(r)
	// ========== 菜单模块初始化 ==========
	InitMenuModule(deps)
	// ========== 注册菜单路由 ==========
	RegisterMenuRoutes(r)
	// ========== 角色模块初始化 ==========
	InitRoleModule(deps)
	// ========== 注册角色路由 ==========
	RegisterRoleRoutes(r)
	// ========== 部门模块初始化 ==========
	InitDeptModule(deps)
	// ========== 注册部门路由 ==========
	RegisterDeptRoutes(r)
	// ========== 数据权限模块初始化 ==========
	InitScopeModule(deps)
	// ========== 注册数据权限路由 ==========
	RegisterScopeRoutes(r)
	// ========== 角色权限关联模块初始化 ==========
	InitRolePermModule(deps)
	// ========== 注册角色权限关联路由 ==========
	RegisterRolePermRoutes(r)
	// ========== 用户角色关联模块初始化 ==========
	InitUserRoleModule(deps)
	// ========== 注册用户角色关联路由 ==========
	RegisterUserRoleRoutes(r)
	// ========== 角色菜单关联模块初始化 ==========
	InitRoleMenuModule(deps)
	// ========== 注册角色菜单关联路由 ==========
	RegisterRoleMenuRoutes(r)
	// ========== 用户部门关联模块初始化 ==========
	InitUserDeptModule(deps)
	// ========== 注册用户部门关联路由 ==========
	RegisterUserDeptRoutes(r)
	// ========== 菜单权限关联模块初始化 ==========
	InitMenuPermModule(deps)
	// ========== 注册菜单权限关联路由 ==========
	RegisterMenuPermRoutes(r)
	// ========== 类目模块初始化 ==========
	InitCategoryModule(deps)
	// ========== 注册类目路由 ==========
	RegisterCategoryRoutes(r)
	// ========== 上传模块初始化 ==========
	InitUploadModule(deps)
	// ========== 注册上传路由 ==========
	RegisterUploadRoutes(r)
	// ========== 商品模块初始化 ==========
	InitProductModule(deps)
	// ========== 注册商品路由 ==========
	RegisterProductRoutes(r)
	// ========== 用户地址管理模块初始化 ==========
	InitAddressModule(deps)
	// ========== 注册用户地址管理路由 ==========
	RegisterAddressRoutes(r)
	// ========== 库存管理模块初始化 ==========
	InitInventoryModule(deps)
	// ========== 注册库存管理路由 ==========
	RegisterInventoryRoutes(r)
	// ========== 用户购物车管理模块初始化 ==========
	InitCartItemModule(deps)
	// ========== 注册用户购物车管理路由 ==========
	RegisterCartItemRoutes(r)
	// ========== 用户订单管理模块初始化 ==========
	InitOrderModule(deps)
	// ========== 注册用户订单管理路由 ==========
	RegisterOrderRoutes(r)
	// ========== 用户支付管理模块初始化 ==========
	InitPaymentModule(deps)
	// ========== 注册用户支付管理路由 ==========
	RegisterPaymentRoutes(r)
	// ========== 操作日志管理模块初始化 ==========
	InitOperationLogModule(deps)
	// ========== 注册操作日志管理路由 ==========
	RegisterOperationLogRoutes(r)
	// ========== 优惠卷管理模块初始化 ==========
	InitCouponModule(deps)
	// ========== 注册优惠卷管理路由 ==========
	RegisterCouponRoutes(r)

	// 静态文件服务 - 上传文件访问
	r.Static("/uploads", "./uploads")

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	r.GET("/api/v1/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
