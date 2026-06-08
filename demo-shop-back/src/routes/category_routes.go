package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var categoryCtrl *handler.CategoryHandler

// InitCategoryModule 类目模块初始化（在InitRoutes中调用）
func InitCategoryModule() {
	categoryCtrl = handler.NewCategoryHandler()
}

// RegisterCategoryRoutes 初始化类目路由
func RegisterCategoryRoutes(r *gin.Engine) {
	categoryGroup := r.Group("/api/v1/platform/category")
	categoryGroup.Use(middleware.AuthMiddleware())
	{
		categoryGroup.POST("", middleware.PermissionMiddleware("platform:category:create"), categoryCtrl.CreateCategory)
		categoryGroup.GET("", middleware.PermissionMiddleware("platform:category:view"), categoryCtrl.GetCategoryList)
		categoryGroup.GET("/:id", middleware.PermissionMiddleware("platform:category:view"), categoryCtrl.GetCategory)
		categoryGroup.GET("/tree", middleware.PermissionMiddleware("platform:category:tree"), categoryCtrl.GetCategoryTree)
		categoryGroup.GET("/children/:id", middleware.PermissionMiddleware("platform:category:children"), categoryCtrl.GetCategoryChildrenList)
		categoryGroup.PUT("/:id", middleware.PermissionMiddleware("platform:category:update"), categoryCtrl.UpdateCategory)
		categoryGroup.DELETE("/:id", middleware.PermissionMiddleware("platform:category:delete"), categoryCtrl.DeleteCategory)
	}
}
