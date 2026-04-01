package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var menuCtrl *handler.MenuHandler

func InitMenuModule() {
	menuRepo := repository.NewMenuRepo()
	menuService := service.NewMenuService(menuRepo)
	menuCtrl = handler.NewMenuHandler(menuService)
}

func RegisterMenuRoutes(r *gin.Engine) {
	permGroup := r.Group("/api/v1/menu")
	permGroup.Use(middleware.AuthMiddleware())
	{
		permGroup.POST("", menuCtrl.CreateMenu)
		permGroup.GET("", menuCtrl.GetMenuList)
		permGroup.GET("/:id", menuCtrl.GetMenu)
		//permGroup.PUT("/:id", menuCtrl.UpdataMenu) // 角色菜单树 角色表暂未构建
		permGroup.PUT("/:id", menuCtrl.UpdataMenu)
		permGroup.DELETE("/:id", menuCtrl.DeleteMenu)
	}
}
