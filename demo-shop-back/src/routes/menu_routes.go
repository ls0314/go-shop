package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var menuCtrl *handler.MenuHandler

// InitMenuModule 菜单模块初始化（在InitRoutes中调用）
func InitMenuModule() {
	menuCtrl = handler.NewMenuHandler()
}

// RegisterMenuRoutes 初始化菜单路由
func RegisterMenuRoutes(r *gin.Engine) {
	menuGroup := r.Group("/api/v1/menu")
	menuGroup.Use(middleware.AuthMiddleware())
	menuGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		menuGroup.POST("", menuCtrl.CreateMenu)
		menuGroup.GET("", menuCtrl.GetMenuList)
		menuGroup.GET("/:id", menuCtrl.GetMenu)
		menuGroup.POST("/tree", menuCtrl.GetMenuTreeByUserId)
		menuGroup.GET("/:id/tree", menuCtrl.GetMenuTreeByRoleId)
		menuGroup.PUT("/:id", menuCtrl.UpdateMenu)
		menuGroup.DELETE("/:id", menuCtrl.DeleteMenu)
	}
}
