package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var orderCtrl *handler.OrderHandler

// InitOrderModule 订单模块初始化（在InitRoutes中调用）
func InitOrderModule(deps service.ServiceDeps) {
	orderCtrl = handler.NewOrderHandler(deps)
}

// RegisterOrderRoutes 初始化订单路由
// 管理端路由前缀：/api/v1/platform/orders，鉴权：AuthMiddleware + PermissionMiddleware
// 用户端路由前缀：/api/v1/user/platform/orders，鉴权：AuthMiddleware（JWT登录即可）
// 管理端所需权限：platform:order:view（列表/详情）/ platform:order:ship（发货）
func RegisterOrderRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(deps.UserRPC, deps.Cache)
	orderGroup := r.Group("/api/v1/admin/orders")
	orderGroup.Use(middleware.AuthMiddleware())
	orderGroup.Use(middleware.OperationLogMiddleware(model.LogModuleOrder))
	{
		// 接口5：管理端订单列表 → GET /api/v1/platform/orders
		orderGroup.GET("", permMW, orderCtrl.GetOrderList)
		// 接口6：管理端订单详情 → GET /api/v1/platform/orders/:id
		orderGroup.GET("/:id", permMW, orderCtrl.GetOrderInfo)
		// 接口7：管理端发货 → PUT /api/v1/platform/orders/:id/ship
		orderGroup.PUT("/:id/ship", permMW, orderCtrl.OrderShip)
	}
	userOrderGroup := r.Group("/api/v1/orders")
	userOrderGroup.Use(middleware.AuthMiddleware())
	{
		// 接口1：创建订单 → POST /api/v1/user/platform/orders
		userOrderGroup.POST("", orderCtrl.CreateOrder)
		// 接口2：用户订单列表 → GET /api/v1/user/platform/orders
		userOrderGroup.GET("", orderCtrl.GetUserOrderList)
		// 接口3：用户订单详情 → GET /api/v1/user/platform/orders/:id
		userOrderGroup.GET("/:id", orderCtrl.GetUserOrderInfo)
		// 接口4：取消订单 → PUT /api/v1/user/platform/orders/:id/cancel
		userOrderGroup.PUT("/:id/cancel", orderCtrl.CancelOrder)
		// 接口8：确认收货 → PUT /api/v1/user/platform/orders/:id/confirm
		userOrderGroup.PUT("/:id/confirm", orderCtrl.ConfirmOrder)
	}
}
