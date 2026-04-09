package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var permCtrl *handler.PermissionHandler

// InitPermissionModule 权限模块初始化（在InitRoutes中调用）
func InitPermissionModule() {
	permRepo := repository.NewPermissionRepo()
	permService := service.NewPermissionService(permRepo)
	permCtrl = handler.NewPermissionHandler(permService)
}

// RegisterPermissionRoutes 初始化权限路由
func RegisterPermissionRoutes(r *gin.Engine) {
	permGroup := r.Group("/api/v1/permissions")
	permGroup.Use(middleware.AuthMiddleware())
	{
		permGroup.POST("", permCtrl.CreatePermission)
		permGroup.GET("", permCtrl.GetPermissionList)
		permGroup.GET("/:id", permCtrl.GetPermission)
		permGroup.PUT("/:id", permCtrl.UpdatePermission)
		permGroup.DELETE("/:id", permCtrl.DeletePermission)
	}
}
