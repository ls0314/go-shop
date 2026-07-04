package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var productCtrl *handler.ProductHandler

// InitProductModule 商品模块初始化（在InitRoutes中调用）
func InitProductModule() {
	productCtrl = handler.NewProductHandler()
}

// RegisterProductRoutes 初始化商品路由
func RegisterProductRoutes(r *gin.Engine) {
	productGroup := r.Group("/api/v1/platform/products")
	productGroup.Use(middleware.AuthMiddleware())
	{
		productGroup.POST("", middleware.PermissionMiddleware("platform:product:create"), productCtrl.CreateProduct)
		productGroup.GET("", middleware.PermissionMiddleware("platform:product:view"), productCtrl.GetProductList)
		productGroup.GET("/:id", middleware.PermissionMiddleware("platform:product:view"), productCtrl.GetProduct)
		productGroup.PUT("/:id", middleware.PermissionMiddleware("platform:product:update"), productCtrl.UpdateProduct)
		productGroup.PUT("/:id/full", middleware.PermissionMiddleware("platform:product:update"), productCtrl.UpdateFullProduct)
		productGroup.POST("/:id/publish", middleware.PermissionMiddleware("platform:product:publish"), productCtrl.PublishProduct)
		productGroup.POST("/:id/withdraw", middleware.PermissionMiddleware("platform:product:withdraw"), productCtrl.WithdrawnProduct)
		productGroup.DELETE("/:id", middleware.PermissionMiddleware("platform:product:delete"), productCtrl.DeleteProduct)
	}

	userProductGroup := r.Group("/api/v1/users/platform/products")
	userProductGroup.Use()
	{
		userProductGroup.GET("", productCtrl.UserProductList)
		userProductGroup.GET(":id", productCtrl.UserProduct)
	}
}
