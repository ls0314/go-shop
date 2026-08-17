package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var couponCtrl *handler.CouponHandler

func InitCouponModule() {

	couponCtrl = handler.NewCouponHandler()

}

func RegisterCouponRoutes(r *gin.Engine) {
	couponGroup := r.Group("/api/v1/admin/platform/coupons")
	couponGroup.Use(middleware.AuthMiddleware())
	couponGroup.Use(middleware.OperationLogMiddleware(model.LogModuleProduct))
	{
		couponGroup.POST("", middleware.PermissionMiddleware(), couponCtrl.CreateCouponTemplate)
		couponGroup.GET("", middleware.PermissionMiddleware(), couponCtrl.GetCouponList)
	}
	userCouponGroup := r.Group("/api/v1/users/platform/coupons")
	userCouponGroup.Use(middleware.AuthMiddleware())
	{
		userCouponGroup.GET("", couponCtrl.GetUserCouponList)
		userCouponGroup.GET("/available", couponCtrl.GetAvailableCouponList)
		userCouponGroup.POST("/receive/:id", couponCtrl.ReceiveCoupon)
	}
}
