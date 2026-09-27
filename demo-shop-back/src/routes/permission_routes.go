package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
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
func RegisterPermissionRoutes(r *gin.Engine) {
	permGroup := r.Group("/api/v1/permissions")
	permGroup.Use(middleware.AuthMiddleware())
	permGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		permGroup.POST("", middleware.PermissionMiddleware(), permCtrl.CreatePermission)
		permGroup.GET("", middleware.PermissionMiddleware(), permCtrl.GetPermissionList)
		permGroup.GET("/:id", middleware.PermissionMiddleware(), permCtrl.GetPermission)
		permGroup.PUT("/:id", middleware.PermissionMiddleware(), permCtrl.UpdatePermission)
		permGroup.DELETE("/:id", middleware.PermissionMiddleware(), permCtrl.DeletePermission)
	}
}
