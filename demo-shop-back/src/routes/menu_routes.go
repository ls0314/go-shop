package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var menuCtrl *handler.MenuHandler

// InitMenuModule 菜单模块初始化（在InitRoutes中调用）
func InitMenuModule(deps service.ServiceDeps) {
	menuCtrl = handler.NewMenuHandler(deps)
}

// RegisterMenuRoutes 初始化菜单路由
func RegisterMenuRoutes(r *gin.Engine) {
	menuGroup := r.Group("/api/v1/menu")
	menuGroup.Use(middleware.AuthMiddleware())
	menuGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		menuGroup.POST("", middleware.PermissionMiddleware(), menuCtrl.CreateMenu)
		menuGroup.GET("", middleware.PermissionMiddleware(), menuCtrl.GetMenuList)
		menuGroup.GET("/:id", middleware.PermissionMiddleware(), menuCtrl.GetMenu)
		menuGroup.POST("/tree", menuCtrl.GetMenuTreeByUserId)
		menuGroup.GET("/:id/tree", middleware.PermissionMiddleware(), menuCtrl.GetMenuTreeByRoleId)
		menuGroup.PUT("/:id", middleware.PermissionMiddleware(), menuCtrl.UpdateMenu)
		menuGroup.DELETE("/:id", middleware.PermissionMiddleware(), menuCtrl.DeleteMenu)
	}
}
