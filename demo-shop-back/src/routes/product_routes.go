package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var productCtrl *handler.ProductHandler

// InitProductModule 商品模块初始化（在InitRoutes中调用）
func InitProductModule(deps service.ServiceDeps) {
	productCtrl = handler.NewProductHandler(deps)
}

// RegisterProductRoutes 初始化商品路由
func RegisterProductRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	productGroup := r.Group("/api/v1/platform/products")
	productGroup.Use(middleware.AuthMiddleware())
	productGroup.Use(middleware.OperationLogMiddleware(model.LogModuleProduct))
	{
		productGroup.POST("", permMW, productCtrl.CreateProduct)
		productGroup.GET("", permMW, productCtrl.GetProductList)
		productGroup.GET("/:id", permMW, productCtrl.GetProduct)
		productGroup.PUT("/:id", permMW, productCtrl.UpdateProduct)
		productGroup.PUT("/:id/full", permMW, productCtrl.UpdateFullProduct)
		productGroup.POST("/:id/publish", permMW, productCtrl.PublishProduct)
		productGroup.POST("/:id/withdraw", permMW, productCtrl.WithdrawnProduct)
		productGroup.DELETE("/:id", permMW, productCtrl.DeleteProduct)
	}

	userProductGroup := r.Group("/api/v1/users/platform/products")
	userProductGroup.Use()
	{
		userProductGroup.GET("", productCtrl.UserProductList)
		userProductGroup.GET(":id", productCtrl.UserProduct)
	}
}
