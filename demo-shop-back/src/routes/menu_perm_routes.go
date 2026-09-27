package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var menuPermissionCtrl *handler.MenuPermissionHandler

// InitMenuPermModule 初始化菜单权限模块
func InitMenuPermModule(deps service.ServiceDeps) {
	menuPermissionCtrl = handler.NewMenuPermissionHandler(deps)
}

// RegisterMenuPermRoutes 注册菜单权限相关路由
func RegisterMenuPermRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	// 菜单权限接口路由分组，统一前缀 /api/v1/menu
	menuPermissionGroup := r.Group("/api/v1/menu")
	// 添加全局认证中间件（必须登录才能访问）
	menuPermissionGroup.Use(middleware.AuthMiddleware())
	menuPermissionGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		// 为菜单分配权限接口
		menuPermissionGroup.POST("/assign-perm", permMW, menuPermissionCtrl.CreateMenuPermissionRel)
		// 根据菜单ID查询关联权限列表接口
		menuPermissionGroup.GET("/:id/perm", permMW, menuPermissionCtrl.GetMenuPermissionRelList)
		// 根据菜单ID清空关联权限接口
		menuPermissionGroup.DELETE("/:id/clear-perm", permMW, menuPermissionCtrl.DeleteMenuAllPermissionRel)
	}
}
