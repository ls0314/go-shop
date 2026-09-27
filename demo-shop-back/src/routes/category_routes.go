package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var categoryCtrl *handler.CategoryHandler

// InitCategoryModule 类目模块初始化（在InitRoutes中调用）
func InitCategoryModule(deps service.ServiceDeps) {
	categoryCtrl = handler.NewCategoryHandler(deps)
}

// RegisterCategoryRoutes 初始化类目路由
func RegisterCategoryRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	categoryGroup := r.Group("/api/v1/platform/category")
	categoryGroup.Use(middleware.AuthMiddleware())
	categoryGroup.Use(middleware.OperationLogMiddleware(model.LogModuleCategory))
	{
		categoryGroup.POST("", permMW, categoryCtrl.CreateCategory)
		categoryGroup.GET("", permMW, categoryCtrl.GetCategoryList)
		categoryGroup.GET("/:id", permMW, categoryCtrl.GetCategory)
		categoryGroup.GET("/tree", permMW, categoryCtrl.GetCategoryTree)
		categoryGroup.GET("/children/:id", permMW, categoryCtrl.GetCategoryChildrenList)
		categoryGroup.PUT("/:id", permMW, categoryCtrl.UpdateCategory)
		categoryGroup.DELETE("/:id", permMW, categoryCtrl.DeleteCategory)
	}
}
