package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var couponCtrl *handler.CouponHandler

// InitCouponModule 初始化优惠券模块 handler 实例
// 说明：在 routes.go 中注册路由前调用，全局单例
func InitCouponModule() {

	couponCtrl = handler.NewCouponHandler()

}

// RegisterCouponRoutes 注册优惠券模块路由
// 管理端（/api/v1/admin/platform/coupons）：创建模板 + 模板列表
//   - AuthMiddleware + OperationLogMiddleware 挂在 group 上
//   - PermissionMiddleware 挂在单条路由上（platform:coupon:create / view）
//
// 用户端（/api/v1/users/platform/coupons）：领取 + 我的券 + 结算可用券
//   - 仅 AuthMiddleware，userId 由 handler 从 JWT 上下文获取
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
		userCouponGroup.GET("/templates", couponCtrl.GetReceiveCouponList)
		userCouponGroup.GET("/available", couponCtrl.GetAvailableCouponList)
		userCouponGroup.POST("/receive/:id", couponCtrl.ReceiveCoupon)
	}
}
