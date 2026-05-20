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
		categoryGroup.POST("", categoryCtrl.CreateCategory)
		categoryGroup.GET("", categoryCtrl.GetCategoryList)
		categoryGroup.GET("/:id", categoryCtrl.GetCategory)
		categoryGroup.GET("/tree", categoryCtrl.GetCategoryTree)
		categoryGroup.GET("/children/:id", categoryCtrl.GetCategoryChildrenList)
		categoryGroup.PUT("/:id", categoryCtrl.UpdateCategory)
		categoryGroup.DELETE("/:id", categoryCtrl.DeleteCategory)
	}
}
