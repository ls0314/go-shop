package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var categoryCtrl *handler.CategoryHandler

// InitCategoryModule 类目模块初始化（在InitRoutes中调用）
func InitCategoryModule(deps service.ServiceDeps) {
	categoryCtrl = handler.NewCategoryHandler(deps)
}

// RegisterCategoryRoutes 初始化类目路由
func RegisterCategoryRoutes(r *gin.Engine) {
	categoryGroup := r.Group("/api/v1/platform/category")
	categoryGroup.Use(middleware.AuthMiddleware())
	categoryGroup.Use(middleware.OperationLogMiddleware(model.LogModuleCategory))
	{
		categoryGroup.POST("", middleware.PermissionMiddleware(), categoryCtrl.CreateCategory)
		categoryGroup.GET("", middleware.PermissionMiddleware(), categoryCtrl.GetCategoryList)
		categoryGroup.GET("/:id", middleware.PermissionMiddleware(), categoryCtrl.GetCategory)
		categoryGroup.GET("/tree", middleware.PermissionMiddleware(), categoryCtrl.GetCategoryTree)
		categoryGroup.GET("/children/:id", middleware.PermissionMiddleware(), categoryCtrl.GetCategoryChildrenList)
		categoryGroup.PUT("/:id", middleware.PermissionMiddleware(), categoryCtrl.UpdateCategory)
		categoryGroup.DELETE("/:id", middleware.PermissionMiddleware(), categoryCtrl.DeleteCategory)
	}
}
