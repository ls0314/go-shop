package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeptHandler handler层部门对象实例
type DeptHandler struct {
	DeptService *service.DeptService // 对齐RoleHandler，使用指针类型
}

// NewDeptHandler 新建handler层部门对象实例
// 接收值：无接收值
// 返回值：*DeptHandler - handler层部门对象指针
func NewDeptHandler(deps service.ServiceDeps) *DeptHandler {
	return &DeptHandler{
		DeptService: service.NewDeptService(deps),
	}
}

// CreateDept 创建部门接口
// 路由映射：POST /api/v1/dept
// 功能：接收前端传递的部门信息，校验参数后调用服务层创建部门并入库
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建部门失败，返回服务器异常信息
//	200：创建成功，返回创建完成的部门完整信息
func (d *DeptHandler) CreateDept(c *gin.Context) {
	// 声明部门实体，用于接收前端请求参数
	var dept model.SysDept
	// 绑定并校验前端传入的JSON/表单参数
	if err := c.ShouldBind(&dept); err != nil {
		// 参数绑定失败，返回400错误响应
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层方法执行部门创建逻辑
	if err := d.DeptService.CreateDept(&dept); err != nil {
		// 服务层执行失败，返回500服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 部门创建成功，返回成功响应与部门数据
	utils.Success(c, dept)
}

// GetDept 根据ID获取单个部门信息接口
// 路由映射：GET /api/v1/dept/:id
// 功能：从URL路径中获取部门ID，查询并返回对应部门详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询部门失败
//	200：查询成功，返回部门详细信息
func (d *DeptHandler) GetDept(c *gin.Context) {
	// 从URL路径参数中获取部门ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist)
		return
	}
	// 调用服务层，根据部门ID查询部门信息
	dept, err := d.DeptService.GetDept(id)
	if err != nil {
		// 服务层查询异常，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回部门信息
	utils.Success(c, dept)
}

// GetDeptTreeByUserId 根据用户Id获取其对应的用户部门树
// 路由映射：GET /api/v1/dept/:userId/tree
// 功能：从URL路径中获取用户ID，查询并返回对应部门树详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询部门树失败
//	200：查询成功，返回部门树详细信息
func (d *DeptHandler) GetDeptTreeByUserId(c *gin.Context) {
	// 从URL路径中获取用户ID
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist)
		return
	}
	// 调用服务层查询对应部门树详情
	deptTree, err := d.DeptService.GetDeptTreeByRoleId(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 返回部门树
	utils.Success(c, deptTree)
}

// GetDeptList 分页获取部门列表接口
// 路由映射：GET /api/v1/dept
// 功能：支持分页、按部门名称/状态筛选查询部门列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取分页参数、筛选条件、返回响应
// 请求参数：
//
//	page     - 页码，默认值 1
//	pageSize - 每页条数，默认值 10
//	deptType - 部门类型（可选筛选条件）
//
// 响应：
//
//	500：服务层查询部门列表失败
//	200：查询成功，返回部门列表、总条数、当前页码、每页条数
func (d *DeptHandler) GetDeptList(c *gin.Context) {
	// 获取页码，未传参则使用默认值 1
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// 获取每页条数，未传参则使用默认值 10
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 获取部门名称筛选条件（可选参数）
	deptType := c.Query("deptType")

	// 调用服务层，分页+条件查询部门列表，返回列表数据和总记录数
	deptList, total, err := d.DeptService.GetDeptList(page, pageSize, deptType)
	if err != nil {
		// 查询失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}

	// 查询成功，返回分页结果
	utils.Success(c, gin.H{
		"list":     deptList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateDept 根据ID更新部门信息接口
// 路由映射：PUT /api/v1/dept/:id
// 功能：从URL获取部门ID，接收前端传入的更新字段，执行部门信息更新，返回更新后的部门详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id         - URL路径参数，部门ID
//	updateDept - 请求体JSON，需要更新的部门字段（map格式）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新部门信息失败 / 查询更新后部门信息失败
//	200：更新成功，返回更新后的完整部门信息
func (d *DeptHandler) UpdateDept(c *gin.Context) {
	// 从URL路径参数中获取部门ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	// 定义map接收前端传入的动态更新字段
	var updateDept map[string]interface{}
	// 绑定请求体JSON参数
	if err := c.ShouldBind(&updateDept); err != nil {
		// 参数绑定失败，返回请求参数错误
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	// 调用服务层执行部门更新操作
	if err := d.DeptService.UpdateDept(id, updateDept); err != nil {
		// 更新失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}

	// 更新成功后，根据ID查询最新的部门信息
	dept, err := d.DeptService.GetDept(id)
	if err != nil {
		// 查询最新部门信息失败，返回服务器错误
		utils.Error(c, 500, err.Error())
		return
	}

	// 更新&查询成功，返回最新部门详情
	utils.Success(c, dept)
}

// DeleteDept 根据ID删除部门接口
// 路由映射：DELETE /api/v1/dept/:id
// 功能：从URL路径获取部门ID，调用服务层执行删除操作，返回删除结果
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除部门失败
//	200：删除成功，返回空数据
func (d *DeptHandler) DeleteDept(c *gin.Context) {
	// 从URL路径参数中获取待删除的部门ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层执行部门删除逻辑
	if err := d.DeptService.DeleteDept(id); err != nil {
		// 删除失败，返回服务器错误响应
		utils.Error(c, 500, err.Error())
		return
	}
	// 删除成功，返回成功响应
	utils.Success(c, nil)
}
