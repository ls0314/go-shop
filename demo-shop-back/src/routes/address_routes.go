package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var addressCtrl *handler.AddressHandler

// InitAddressModule 用户地址模块初始化（在InitRoutes中调用）
func InitAddressModule() {
	addressCtrl = handler.NewAddressHandler()
}

// RegisterAddressRoutes 初始化用户地址路由
func RegisterAddressRoutes(r *gin.Engine) {
	addressGroup := r.Group("/api/v1/user/addresses")
	addressGroup.Use(middleware.AuthMiddleware())
	{
		addressGroup.POST("", addressCtrl.CreateAddress)
		addressGroup.GET("", addressCtrl.GetAddressList)
		addressGroup.GET("/:id", addressCtrl.GetAddress)
		addressGroup.PUT("/:id", addressCtrl.UpdateAddress)
		addressGroup.DELETE("/:id", addressCtrl.DeleteAddress)
		addressGroup.PUT("/:id/default", addressCtrl.SetDefaultAddress)
	}
}
