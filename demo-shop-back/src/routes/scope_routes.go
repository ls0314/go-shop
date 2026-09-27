package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

// scopeCtrl 数据权限模块全局控制器实例
// 作用：供路由注册时使用，持有部门Handler单例对象
var scopeCtrl *handler.ScopeHandler

// InitScopeModule 初始化数据权限模块
// 功能：完成数据权限模块 仓库层 → 服务层 → 控制层 的依赖注入与实例化
// 执行顺序：创建数据访问层实例 → 创建业务逻辑层实例 → 创建控制器实例
func InitScopeModule() {
	// 初始化数据权限服务层
	scopeService := service.NewScopeService()
	// 初始化数据权限控制器，赋值给全局控制器变量
	scopeCtrl = handler.NewScopeHandler(scopeService)
}

// RegisterScopeRoutes 注册数据权限模块路由
// 参数：r *gin.Engine Gin路由引擎实例
// 功能：注册数据权限相关API路由，统一前缀 /api/v1/scope，并添加登录认证中间件
func RegisterScopeRoutes(r *gin.Engine) {
	// 创建数据权限接口路由分组，统一前缀 /api/v1/scope
	scopeGroup := r.Group("/api/v1/scope")
	// 添加全局认证中间件（必须登录才能访问数据权限接口）
	scopeGroup.Use(middleware.AuthMiddleware())
	{
		// 创建数据权限接口
		scopeGroup.POST("", middleware.PermissionMiddleware(), scopeCtrl.CreateScope)
		// 分页获取数据权限列表接口
		scopeGroup.GET("", middleware.PermissionMiddleware(), scopeCtrl.GetScopeList)
		// 根据ID获取单个数据权限接口
		scopeGroup.GET("/:id", middleware.PermissionMiddleware(), scopeCtrl.GetScope)
		// 根据ID更新数据权限接口
		scopeGroup.PUT("/:id", middleware.PermissionMiddleware(), scopeCtrl.UpdateScope)
		// 根据ID删除数据权限接口
		scopeGroup.DELETE("/:id", middleware.PermissionMiddleware(), scopeCtrl.DeleteScope)
	}
}
