package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

// payCtrl 支付模块handler全局实例（在InitPaymentModule中初始化）
var payCtrl *handler.PaymentHandler

// InitPaymentModule 支付模块初始化（在InitRoutes中调用）
func InitPaymentModule(deps service.ServiceDeps) {
	payCtrl = handler.NewPaymentHandler(deps)
}

// RegisterPaymentRoutes 注册支付模块路由
// 路由前缀说明：
//
//	用户端  /api/v1/users/pay       - 鉴权：AuthMiddleware（JWT登录即可）
//	回调    /api/v1/pay              - 鉴权：无（支付平台回调/模拟支付均无需登录）
//	管理端  共用用户端路由前缀          - 鉴权：AuthMiddleware + PermissionMiddleware（platform:pay:view）
func RegisterPaymentRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(deps.UserRPC, deps.Cache)
	// ============================================================
	// 用户端 + 管理端路由（需JWT登录）
	// ============================================================
	payGroup := r.Group("/api/v1/pay")
	payGroup.Use(middleware.AuthMiddleware())
	{
		// 发起支付 → POST /api/v1/user/pay/order/:orderId
		// 鉴权：登录即可，无额外权限要求
		// 参数：路径参数 orderId（订单ID），Body {pay_method}
		payGroup.POST("/order/:id", payCtrl.CreatePayment)

		// 查询支付状态 → GET /api/v1/user/pay/:payNo
		// 鉴权：登录即可，校验支付记录归属
		// 参数：路径参数 payNo（支付流水号）
		payGroup.GET("/:payNo", payCtrl.GetPayment)

	}
	adminPayGroup := r.Group("/api/v1/admin/pay")
	adminPayGroup.Use(middleware.AuthMiddleware())
	{
		// 管理端支付列表 → GET /api/v1/admin/pay/list
		// 鉴权：platform:pay:view（仅超级管理员/运营人员）
		// 参数：Query参数 page/page_size/pay_status/pay_method/order_no/start_time/end_time
		adminPayGroup.GET("/list", permMW, payCtrl.GetPaymentList)
	}

	// ============================================================
	// 支付回调路由（无需鉴权 — 由支付网关签名保证安全）
	// ============================================================
	mockPayGroup := r.Group("/api/v1/pay")
	mockPayGroup.Use()
	{
		// 模拟支付回调 → POST /api/v1/pay/callback/mock
		// 鉴权：无（正式环境由支付平台异步通知，mock环境由前端主动调用）
		// 参数：Body {pay_no, trade_no(可选)}
		mockPayGroup.POST("/callback/mock", payCtrl.MockCallback)
	}
}
