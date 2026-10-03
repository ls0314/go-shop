package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InventoryHandler 库存管理handler层实例
type InventoryHandler struct {
	InventoryService *service.InventoryService // 库存服务层对象指针
}

// NewInventoryHandler 创建库存管理handler层实例
// 接收值：无接收值，全局实例化
// 返回值：*InventoryHandler - 库存handler指针
func NewInventoryHandler(deps service.ServiceDeps) *InventoryHandler {
	return &InventoryHandler{
		InventoryService: service.NewInventoryService(deps),
	}
}

// failRPC 统一处理 RPC 错误(定义见 product_handler.go):
// 下游不可用回 503,其余回 500 —— 库存查询读路径已迁 product-service。

// GetSkuStock 查询单个SKU库存接口
// 路由映射：GET /api/v1/admin/inventory/sku/:id
// 功能：从URL路径获取SKU ID，查询并返回SKU完整库存信息（含锁定库存、销量）
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询SKU库存失败
//	200：查询成功，返回SKU库存完整信息
func (ih *InventoryHandler) GetSkuStock(c *gin.Context) {
	// 通过传入URL地址获取INT格式的SKU ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层查询SKU库存信息
	skuStock, err := ih.InventoryService.GetSkuStock(id)
	if err != nil {
		failRPC(c, err)
		return
	}
	// 查询成功，返回SKU库存信息
	utils.Success(c, skuStock)
}

// GetSkuListBySpu 查询SPU下所有SKU库存接口
// 路由映射：GET /api/v1/admin/inventory/spu/:id
// 功能：从URL路径获取SPU ID，查询该SPU下所有未删除SKU的库存汇总（含各SKU明细和总量）
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询SPU库存列表失败
//	200：查询成功，返回SKU库存列表和汇总数据
func (ih *InventoryHandler) GetSkuListBySpu(c *gin.Context) {
	// 通过传入URL地址获取INT格式的SPU ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层查询SPU下所有SKU库存信息
	skuStockList, err := ih.InventoryService.GetSkuStockListBySpu(id)
	if err != nil {
		failRPC(c, err)
		return
	}
	// 查询成功，返回SKU库存列表
	utils.Success(c, skuStockList)
}

// AdjustStock 手动调整库存接口
// 路由映射：POST /api/v1/admin/inventory/adjust
// 功能：接收前端传入的SKU ID、变更数量和原因，校验参数后调用服务层在事务内执行库存调整并记录日志
// 参数：c *gin.Context Gin上下文，用于获取当前用户、接收请求体、返回响应
// 请求参数：
//
//	sku_id     - SKU ID，int64类型，必填
//	change_qty - 变更数量，int64类型，正数增加负数减少，必填
//	remark     - 调整原因，string类型，必填
//
// 响应：
//
//	400：请求参数绑定失败/remark为空
//	500：服务层调整库存失败（库存不足/SKU不存在等）
//	200：调整成功，返回调整前后库存数量
func (ih *InventoryHandler) AdjustStock(c *gin.Context) {
	// 实例化后绑定请求参数
	var req requset.InventoryAdjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前操作用户ID
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层执行库存调整
	adjustStock, err := ih.InventoryService.AdjustStock(userId, req)
	if err != nil {
		failRPC(c, err)
		return
	}
	// 调整成功，返回调整前后库存数据
	utils.Success(c, adjustStock)
}

// GetStockLog 查询库存变更日志接口
// 路由映射：GET /api/v1/admin/inventory/log
// 功能：支持分页、按SKU/SPU/变更类型/时间范围筛选查询库存变更日志，关联查询SKU名称和SPU名称
// 参数：c *gin.Context Gin上下文，用于获取分页参数、筛选条件、返回响应
// 查询参数：
//
//	page        - 页码，int类型，默认1
//	pageSize    - 每页条数，int类型，默认50
//	sku_id      - SKU筛选，int64类型，可选
//	spu_id      - SPU筛选，int64类型，可选
//	change_type - 变更类型筛选，string类型，可选
//	start_time  - 开始时间，string类型（RFC3339），可选
//	end_time    - 结束时间，string类型（RFC3339），可选
//
// 响应：
//
//	400：请求参数绑定失败
//	500：服务层查询日志失败
//	200：查询成功，返回日志列表和分页信息
func (ih *InventoryHandler) GetStockLog(c *gin.Context) {
	// 实例化后绑定查询参数
	var req requset.InventoryLogReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 调用服务层分页查询库存变更日志
	stockLogList, err := ih.InventoryService.GetStockLogList(req)
	if err != nil {
		failRPC(c, err)
		return
	}
	// 查询成功，返回日志列表和分页信息
	utils.Success(c, stockLogList)
}

// GetWarnStockList 查询低库存预警列表接口
// 路由映射：GET /api/v1/admin/inventory/warning
// 功能：查询可用库存低于阈值的SKU列表，支持自定义阈值和SPU状态筛选，按库存从低到高排列
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 查询参数：
//
//	threshold  - 预警阈值，int类型，默认10
//	spu_status - SPU状态筛选，string类型，默认"published"
//
// 响应：
//
//	400：请求参数绑定失败
//	500：服务层查询低库存预警失败
//	200：查询成功，返回低库存SKU数组
func (ih *InventoryHandler) GetWarnStockList(c *gin.Context) {
	// 实例化后绑定查询参数
	var req requset.InventoryWarnReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 调用服务层查询低库存预警列表
	warnStockList, err := ih.InventoryService.GetWarnStockList(req)
	if err != nil {
		failRPC(c, err)
		return
	}
	// 查询成功，返回预警列表
	utils.Success(c, warnStockList)
}
