package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var inventoryCtrl *handler.InventoryHandler

// InitInventoryModule 库存模块初始化（在InitRoutes中调用）
func InitInventoryModule() {
	inventoryCtrl = handler.NewInventoryHandler()
}

// RegisterInventoryRoutes 初始化库存路由
// 路由前缀：/api/v1/admin/inventory
// 鉴权方式：AuthMiddleware + PermissionMiddleware（按api_path动态匹配权限编码）
// 权限要求：平台超级管理员(全权限) / 平台运营人员(仅日志查看)
func RegisterInventoryRoutes(r *gin.Engine) {
	inventoryGroup := r.Group("/api/v1/admin/inventory")
	inventoryGroup.Use(middleware.AuthMiddleware())
	inventoryGroup.Use(middleware.OperationLogMiddleware(model.LogModuleInventory))
	{
		// 接口1：查询单个SKU库存 → GET /api/v1/admin/inventory/sku/:id
		inventoryGroup.GET("/sku/:id", middleware.PermissionMiddleware(), inventoryCtrl.GetSkuStock)
		// 接口2：查询SPU下所有SKU库存 → GET /api/v1/admin/inventory/spu/:id
		inventoryGroup.GET("/spu/:id", middleware.PermissionMiddleware(), inventoryCtrl.GetSkuListBySpu)
		// 接口3：手动调整库存 → POST /api/v1/admin/inventory/adjust
		inventoryGroup.POST("/adjust", middleware.PermissionMiddleware(), inventoryCtrl.AdjustStock)
		// 接口4：查询库存变更日志 → GET /api/v1/admin/inventory/log
		inventoryGroup.GET("/log", middleware.PermissionMiddleware(), inventoryCtrl.GetStockLog)
		// 接口5：查询低库存预警列表 → GET /api/v1/admin/inventory/warning
		inventoryGroup.GET("/warning", middleware.PermissionMiddleware(), inventoryCtrl.GetWarnStockList)
	}
}
