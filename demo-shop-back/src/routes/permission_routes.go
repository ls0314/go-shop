package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var permCtrl *handler.PermissionHandler

// InitPermissionModule 权限模块初始化（在InitRoutes中调用）
func InitPermissionModule(deps service.ServiceDeps) {
	permService := service.NewPermissionService(deps)
	permCtrl = handler.NewPermissionHandler(permService)
}

// RegisterPermissionRoutes 初始化权限路由
func RegisterPermissionRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	permGroup := r.Group("/api/v1/admin/permissions")
	permGroup.Use(middleware.AuthMiddleware())
	permGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		permGroup.POST("", permMW, permCtrl.CreatePermission)
		permGroup.GET("", permMW, permCtrl.GetPermissionList)
		permGroup.GET("/:id", permMW, permCtrl.GetPermission)
		permGroup.PUT("/:id", permMW, permCtrl.UpdatePermission)
		permGroup.DELETE("/:id", permMW, permCtrl.DeletePermission)
	}
}
