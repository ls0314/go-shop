package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var cartItemCtrl *handler.CartItemHandler

// InitCartItemModule 购物车模块初始化（在InitRoutes中调用）
func InitCartItemModule(deps service.ServiceDeps) {
	cartItemCtrl = handler.NewCartItemHandler(deps)
}

// RegisterCartItemRoutes 初始化购物车路由
// 路由前缀：/api/v1/user/cart
// 鉴权方式：AuthMiddleware（JWT登录即可，无需额外权限）
// 业务规则：归属校验（用户仅能操作自己的购物车）、重复SKU累加、实时联表查询价格库存
func RegisterCartItemRoutes(r *gin.Engine, deps service.ServiceDeps) {
	cartItemGroup := r.Group("/api/v1/users/cart")
	cartItemGroup.Use(middleware.AuthMiddleware())
	{
		// 接口1：加入购物车 → POST /api/v1/user/cart
		cartItemGroup.POST("", cartItemCtrl.AddCartItem)
		// 接口2：获取购物车列表 → GET /api/v1/user/cart
		cartItemGroup.GET("", cartItemCtrl.GetCartItemList)
		// 接口6：获取购物车总数量 → GET /api/v1/user/cart/count
		cartItemGroup.GET("/count", cartItemCtrl.GetCartItemTotal)
		// 接口7：选中项结算预览 → GET /api/v1/user/cart/preview
		cartItemGroup.GET("/preview", cartItemCtrl.GetPayPreviewCartItem)
		// 接口3：修改购物车项（数量/选中） → PUT /api/v1/user/cart/:id
		cartItemGroup.PUT("/:id", cartItemCtrl.UpdateCartItem)
		// 接口5：全选/取消全选 → PUT /api/v1/user/cart/select-all
		cartItemGroup.PUT("/select-all", cartItemCtrl.SelectAllCartItem)
		// 接口4：删除购物车项 → DELETE /api/v1/user/cart/:id
		cartItemGroup.DELETE("/:id", cartItemCtrl.DeleteCartItem)
	}
}
