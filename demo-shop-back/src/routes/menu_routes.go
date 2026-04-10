package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var menuCtrl *handler.MenuHandler

// InitMenuModule 菜单模块初始化（在InitRoutes中调用）
func InitMenuModule() {
	menuRepo := repository.NewMenuRepo()
	menuService := service.NewMenuService(menuRepo)
	menuCtrl = handler.NewMenuHandler(menuService)
}

// RegisterMenuRoutes 初始化菜单路由
func RegisterMenuRoutes(r *gin.Engine) {
	menuGroup := r.Group("/api/v1/menu")
	menuGroup.Use(middleware.AuthMiddleware())
	{
		menuGroup.POST("", menuCtrl.CreateMenu)
		menuGroup.GET("", menuCtrl.GetMenuList)
		menuGroup.GET("/:id", menuCtrl.GetMenu)
		// TODO： 构建角色菜单接口映射
		//permGroup.GET("/tree/:id", menuCtrl.UpdataMenu) // 角色菜单树 角色表暂未构建
		menuGroup.PUT("/:id", menuCtrl.UpdateMenu)
		menuGroup.DELETE("/:id", menuCtrl.DeleteMenu)
	}
}
