package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleHandler handler层角色对象实例
type RoleHandler struct {
	RoleService *service.RoleService
}

// NewRoleHandler 新建handler层角色对象实例
// 接收值：roleService - 服务层角色对象指针
// 返回值：*RoleHandler - handler层角色对象指针
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{RoleService: roleService}
}

// CreateRole 创建角色接口
// 路由映射：POST /api/v1/role
// 功能：接收前端传递的角色信息，校验参数后调用服务层创建角色并入库
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建角色失败，返回服务器异常信息
//	200：创建成功，返回创建完成的角色完整信息
func (r *RoleHandler) CreateRole(c *gin.Context) {
	// 声明角色实体，用于接收前端请求参数
	var role model.SysRole
	// 绑定并校验前端传入的JSON/表单参数
	if err := c.ShouldBind(&role); err != nil {
		// 参数绑定失败，返回400错误响应
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层方法执行角色创建逻辑
	if err := r.RoleService.CreateRole(&role); err != nil {
		// 服务层执行失败，返回500服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 角色创建成功，返回成功响应与角色数据
	utils.Success(c, role)
}

// GetRole 根据ID获取单个角色信息接口
// 路由映射：GET /api/v1/role/:id
// 功能：从URL路径中获取角色ID，查询并返回对应角色详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询角色失败
//	200：查询成功，返回角色详细信息
func (r *RoleHandler) GetRole(c *gin.Context) {
	// 从URL路径参数中获取角色ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层，根据角色ID查询角色信息
	role, err := r.RoleService.GetRoleById(id)
	if err != nil {
		// 服务层查询异常，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回角色信息
	utils.Success(c, role)
}

// GetRoleList 分页获取角色列表接口
// 路由映射：GET /api/v1/role
// 功能：支持分页、按角色类型筛选查询角色列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取分页参数、筛选条件、返回响应
// 请求参数：
//
//	page     - 页码，默认值 1
//	pageSize - 每页条数，默认值 10
//	roleType - 角色类型（可选筛选条件）
//
// 响应：
//
//	500：服务层查询角色列表失败
//	200：查询成功，返回角色列表、总条数、当前页码、每页条数
func (r *RoleHandler) GetRoleList(c *gin.Context) {
	// 获取页码，未传参则使用默认值 1
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// 获取每页条数，未传参则使用默认值 10
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 获取角色类型筛选条件（可选参数）
	roleType := c.Query("roleType")

	// 调用服务层，分页+条件查询角色列表，返回列表数据和总记录数
	roleList, total, err := r.RoleService.GetRoleList(page, pageSize, roleType)
	if err != nil {
		// 查询失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}

	// 查询成功，返回分页结果
	utils.Success(c, gin.H{
		"list":     roleList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateRole 根据ID更新角色信息接口
// 路由映射：PUT /api/v1/role/:id
// 功能：从URL获取角色ID，接收前端传入的更新字段，执行角色信息更新，返回更新后的角色详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id         - URL路径参数，角色ID
//	updateRole - 请求体JSON，需要更新的角色字段（map格式）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新角色信息失败 / 查询更新后角色信息失败
//	200：更新成功，返回更新后的完整角色信息
func (r *RoleHandler) UpdateRole(c *gin.Context) {
	// 从URL路径参数中获取角色ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	// 定义map接收前端传入的动态更新字段
	var updateRole map[string]interface{}
	// 绑定请求体JSON参数
	if err := c.ShouldBind(&updateRole); err != nil {
		// 参数绑定失败，返回请求参数错误
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	// 调用服务层执行角色更新操作
	if err := r.RoleService.UpdateRole(id, updateRole); err != nil {
		// 更新失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}

	// 更新成功后，根据ID查询最新的角色信息
	role, err := r.RoleService.GetRoleById(id)
	if err != nil {
		// 查询最新角色信息失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}

	// 更新&查询成功，返回最新角色详情
	utils.Success(c, role)
}

// DeleteRole 根据ID删除角色接口
// 路由映射：DELETE /api/role/:id (具体路由以实际注册为准)
// 功能：从URL路径获取角色ID，调用服务层执行删除操作，返回删除结果
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除角色失败
//	200：删除成功，返回空数据
func (r *RoleHandler) DeleteRole(c *gin.Context) {
	// 从URL路径参数中获取待删除的角色ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层执行角色删除逻辑
	if err := r.RoleService.DeleteRole(id); err != nil {
		// 删除失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}
	// 删除成功，返回成功响应
	utils.Success(c, nil)
}
