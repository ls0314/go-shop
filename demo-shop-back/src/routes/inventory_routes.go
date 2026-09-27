package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var inventoryCtrl *handler.InventoryHandler

// InitInventoryModule 库存模块初始化（在InitRoutes中调用）
func InitInventoryModule(deps service.ServiceDeps) {
	inventoryCtrl = handler.NewInventoryHandler(deps)
}

// RegisterInventoryRoutes 初始化库存路由
// 路由前缀：/api/v1/admin/inventory
// 鉴权方式：AuthMiddleware + PermissionMiddleware（按api_path动态匹配权限编码）
// 权限要求：平台超级管理员(全权限) / 平台运营人员(仅日志查看)
func RegisterInventoryRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	inventoryGroup := r.Group("/api/v1/admin/inventory")
	inventoryGroup.Use(middleware.AuthMiddleware())
	inventoryGroup.Use(middleware.OperationLogMiddleware(model.LogModuleInventory))
	{
		// 接口1：查询单个SKU库存 → GET /api/v1/admin/inventory/sku/:id
		inventoryGroup.GET("/sku/:id", permMW, inventoryCtrl.GetSkuStock)
		// 接口2：查询SPU下所有SKU库存 → GET /api/v1/admin/inventory/spu/:id
		inventoryGroup.GET("/spu/:id", permMW, inventoryCtrl.GetSkuListBySpu)
		// 接口3：手动调整库存 → POST /api/v1/admin/inventory/adjust
		inventoryGroup.POST("/adjust", permMW, inventoryCtrl.AdjustStock)
		// 接口4：查询库存变更日志 → GET /api/v1/admin/inventory/log
		inventoryGroup.GET("/log", permMW, inventoryCtrl.GetStockLog)
		// 接口5：查询低库存预警列表 → GET /api/v1/admin/inventory/warning
		inventoryGroup.GET("/warning", permMW, inventoryCtrl.GetWarnStockList)
	}
}
