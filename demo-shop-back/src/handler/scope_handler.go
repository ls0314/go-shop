package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ScopeHandler handler层权限范围对象实例
type ScopeHandler struct {
	ScopeService *service.ScopeService // 服务层权限范围对象指针
}

// NewScopeHandler 新建handler层权限范围对象实例
// 接收值：scopeService - 服务层权限范围对象指针
// 返回值：*ScopeHandler - handler层权限范围对象指针
func NewScopeHandler(scopeService *service.ScopeService) *ScopeHandler {
	return &ScopeHandler{ScopeService: scopeService}
}

// CreateScope 创建权限范围接口
// 路由映射：POST /api/v1/scope
// 功能：接收前端传递的权限范围信息，校验参数后调用服务层创建权限范围并入库
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建权限范围失败，返回服务器异常信息
//	200：创建成功，返回创建完成的权限范围完整信息
func (s *ScopeHandler) CreateScope(c *gin.Context) {
	// 声明权限范围实体指针，用于接收前端请求参数
	var scope *model.SysScope
	// 绑定并校验前端传入的JSON/表单参数
	if err := c.ShouldBind(&scope); err != nil {
		// 参数绑定失败，返回400错误响应
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层方法执行权限范围创建逻辑
	if err := s.ScopeService.CreateScope(scope); err != nil {
		// 服务层执行失败，返回500服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 权限范围创建成功，返回成功响应与权限范围数据
	utils.Success(c, scope)
}

// GetScope 根据ID获取单个权限范围信息接口
// 路由映射：GET /api/v1/scope/:id
// 功能：从URL路径中获取权限范围ID，查询并返回对应权限范围详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询权限范围失败
//	200：查询成功，返回权限范围详细信息
func (s *ScopeHandler) GetScope(c *gin.Context) {
	// 从URL路径参数中获取权限范围ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层，根据权限范围ID查询权限范围信息
	scope, err := s.ScopeService.GetScopeById(id)
	if err != nil {
		// 服务层查询异常，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回权限范围信息
	utils.Success(c, scope)

}

// GetScopeList 分页获取权限范围列表接口
// 路由映射：GET /api/v1/scope
// 功能：支持分页、按资源类型筛选查询权限范围列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取分页参数、筛选条件、返回响应
// 请求参数：
//
//	page         - 页码，默认值 1
//	pageSize     - 每页条数，默认值 10
//	resourceType - 资源类型（可选筛选条件）
//
// 响应：
//
//	500：服务层查询权限范围列表失败
//	200：查询成功，返回权限范围列表、总条数、当前页码、每页条数
func (s *ScopeHandler) GetScopeList(c *gin.Context) {
	// 获取页码，未传参则使用默认值 1
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// 获取每页条数，未传参则使用默认值 10
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 获取资源类型筛选条件（可选参数）
	resourceType, _ := c.GetQuery("resourceType")

	// 调用服务层，分页+条件查询权限范围列表，返回列表数据和总记录数
	scopeList, total, err := s.ScopeService.GetScopeList(page, pageSize, resourceType)
	if err != nil {
		// 查询失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回分页结果
	utils.Success(c, gin.H{
		"list":     scopeList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateScope 根据ID更新权限范围信息接口
// 路由映射：PUT /api/v1/scope/:id
// 功能：从URL获取权限范围ID，接收前端传入的更新字段，执行权限范围信息更新，返回更新后的权限范围详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id          - URL路径参数，权限范围ID
//	updateScope - 请求体JSON，需要更新的权限范围字段（map格式）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新权限范围信息失败 / 查询更新后权限范围信息失败
//	200：更新成功，返回更新后的完整权限范围信息
func (s *ScopeHandler) UpdateScope(c *gin.Context) {
	// 从URL路径参数中获取权限范围ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	// 定义map接收前端传入的动态更新字段
	var updateScope map[string]interface{}
	// 绑定请求体JSON参数
	if err := c.ShouldBind(&updateScope); err != nil {
		// 参数绑定失败，返回请求参数错误
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	// 调用服务层执行权限范围更新操作
	if err := s.ScopeService.UpdateScope(id, updateScope); err != nil {
		// 更新失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 更新成功后，根据ID查询最新的权限范围信息
	scope, err := s.ScopeService.GetScopeById(id)
	if err != nil {
		// 查询最新权限范围信息失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 更新&查询成功，返回最新权限范围详情
	utils.Success(c, scope)
}

// DeleteScope 根据ID删除权限范围接口
// 路由映射：DELETE /api/v1/scope/:id
// 功能：从URL路径获取权限范围ID，调用服务层执行删除操作，返回删除结果
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除权限范围失败
//	200：删除成功，返回空数据
func (s *ScopeHandler) DeleteScope(c *gin.Context) {
	// 从URL路径参数中获取待删除的权限范围ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层执行权限范围删除逻辑
	if err := s.ScopeService.DeleteScope(id); err != nil {
		// 删除失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}
	// 删除成功，返回成功响应
	utils.Success(c, nil)
}
