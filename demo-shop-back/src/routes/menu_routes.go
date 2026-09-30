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
	menuCtrl = handler.NewMenuHandler(deps.UserRPC)
}

// RegisterMenuRoutes 初始化菜单路由
func RegisterMenuRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(deps.UserRPC, deps.Cache)
	menuGroup := r.Group("/api/v1/admin/menu")
	menuGroup.Use(middleware.AuthMiddleware())
	menuGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		menuGroup.POST("", permMW, menuCtrl.CreateMenu)
		menuGroup.GET("", permMW, menuCtrl.GetMenuList)
		menuGroup.GET("/:id", permMW, menuCtrl.GetMenu)
		menuGroup.POST("/tree", menuCtrl.GetMenuTreeByUserId)
		menuGroup.GET("/:id/tree", permMW, menuCtrl.GetMenuTreeByRoleId)
		menuGroup.PUT("/:id", permMW, menuCtrl.UpdateMenu)
		menuGroup.DELETE("/:id", permMW, menuCtrl.DeleteMenu)
	}
}
