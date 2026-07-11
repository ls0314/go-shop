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
// 路由前缀：/api/v1/user/addresses
// 鉴权方式：AuthMiddleware（JWT登录即可，无需额外权限）
// 业务规则：归属校验（用户仅能操作自己的地址）、默认地址管理（唯一默认+自动转移）、上限20条
func RegisterAddressRoutes(r *gin.Engine) {
	addressGroup := r.Group("/api/v1/user/addresses")
	addressGroup.Use(middleware.AuthMiddleware())
	{
		// 接口1：新增收货地址 → POST /api/v1/user/addresses
		addressGroup.POST("", addressCtrl.CreateAddress)
		// 接口2：获取用户地址列表 → GET /api/v1/user/addresses
		addressGroup.GET("", addressCtrl.GetAddressList)
		// 接口3：获取单个地址详情 → GET /api/v1/user/addresses/:id
		addressGroup.GET("/:id", addressCtrl.GetAddress)
		// 接口4：更新收货地址 → PUT /api/v1/user/addresses/:id
		addressGroup.PUT("/:id", addressCtrl.UpdateAddress)
		// 接口5：删除收货地址（软删除） → DELETE /api/v1/user/addresses/:id
		addressGroup.DELETE("/:id", addressCtrl.DeleteAddress)
		// 接口6：设为默认地址 → PUT /api/v1/user/addresses/:id/default
		addressGroup.PUT("/:id/default", addressCtrl.SetDefaultAddress)
	}
}
