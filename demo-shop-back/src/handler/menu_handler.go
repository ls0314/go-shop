package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单表handler层实例
type MenuHandler struct {
	MenuService *service.MenuService // 菜单服务层对象指针
}

// NewMenuHandler 新建菜单表中的HTTP handler实例
// 接收值：无接收值，全局实例化
// 返回值：*MenuHandler - 菜单handler指针
func NewMenuHandler() *MenuHandler {
	return &MenuHandler{
		MenuService: service.NewMenuService(),
	}
}

// CreateMenu 创建菜单接口
// 路由映射：POST /api/v1/menu
// 功能：接收前端传递的菜单信息，校验参数后调用服务层创建菜单并入库
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建菜单失败，返回服务器异常信息
//	200：创建成功，返回创建完成的菜单完整信息
func (m *MenuHandler) CreateMenu(c *gin.Context) {
	// 实例化后绑定参数
	var menu model.SysMenu
	if err := c.ShouldBind(&menu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层创建菜单
	if err := m.MenuService.CreateMenu(&menu); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 创建成功，返回新建菜单数据
	utils.Success(c, menu)
}

// GetMenu 查询菜单信息接口（根据ID查询）
// 路由映射：GET /api/v1/menu/:id
// 功能：从URL路径中获取菜单ID，查询并返回对应菜单详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询菜单失败
//	200：查询成功，返回菜单详细信息
func (m *MenuHandler) GetMenu(c *gin.Context) {
	// 通过传入URL地址获取INT格式的ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层查找权限信息
	menu, err := m.MenuService.GetMenu(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回所查询菜单信息
	utils.Success(c, menu)
}

// GetMenuTreeByRoleId 根据角色ID获取对应的菜单树
// 路由映射：GET /api/v1/menu/:roleId/tree
// 功能：从URL路径中获取角色ID，查询并返回对应菜单树详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询菜单树失败
//	200：查询成功，返回菜单树详细信息
func (m *MenuHandler) GetMenuTreeByRoleId(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	var roleIds []int64
	roleIds = append(roleIds, roleId)
	menuTree, err := m.MenuService.GetMenuTreeByRoleIds(roleIds)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, menuTree)
}

// GetMenuTreeByUserId 查询用户菜单树
// 路由映射：POST /api/v1/menu/tree
// 功能： 根据用户ID查询用户对应的全部的角色的关联菜单，并返回并集菜单树
// 参数： ：c *gin.Context Gin上下文，用于获取参数
// 请求参数：
//
//	userID - 用户ID
//
// 响应：
//
//	500：服务层查询失败
//	200：查询成功，返回角色并集菜单树
//	400：请求参数错误
func (m *MenuHandler) GetMenuTreeByUserId(c *gin.Context) {
	// 定义请求结构体
	type UserIdRequest struct {
		UserId int64 `json:"userId" binding:"required"`
	}

	// 绑定到结构体
	var req UserIdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层查询并集菜单树
	menuTree, err := m.MenuService.GetMenuTreeByUserId(req.UserId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, menuTree)
}

// GetMenuList 分页查询菜单信息接口
// 路由映射：GET /api/v1/menu
// 功能：支持分页、按菜单类型筛选查询菜单列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取分页参数、筛选条件、返回响应
// 请求参数：
//
//	page     - 页码，默认值 1
//	pageSize - 每页条数，默认值 10
//	menuType - 菜单类型（可选筛选条件）
//
// 响应：
//
//	500：服务层查询菜单列表失败
//	200：查询成功，返回菜单列表、总条数、当前页码、每页条数
func (m *MenuHandler) GetMenuList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	permType := c.Query("menuType")
	// 调用服务层分页查询菜单
	menu, total, err := m.MenuService.GetMenuList(page, pageSize, permType)
	if err != nil {
		utils.Error(c, 500, "查询失败")
		return
	}

	// 返回分页数据
	utils.Success(c, gin.H{
		"list":     menu,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateMenu 更新菜单接口
// 路由映射：PUT /api/v1/menu/:id
// 功能：从URL获取菜单ID，接收前端传入的更新字段，执行菜单信息更新，返回更新后的菜单详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id         - URL路径参数，菜单ID
//	updateMenu - 请求体JSON，需要更新的菜单字段（map格式）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新菜单信息失败 / 查询更新后菜单信息失败
//	200：更新成功，返回更新后的完整菜单信息
func (m *MenuHandler) UpdateMenu(c *gin.Context) {
	// 通过URL地址获取所更新菜单ID的INT格式（json传进格式一般为float64，故此处从URL取id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 实例化后绑定参数, 此处使用map来确保菜单信息可以局部更新
	var updateMenu map[string]interface{}
	if err := c.ShouldBind(&updateMenu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	//调用服务层更新菜单部分信息
	if err := m.MenuService.UpdateMenu(id, updateMenu); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 获取菜单更新后的完整信息
	menu, err := m.MenuService.GetMenu(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
	}
	// 更新成功,返回更新后完整菜单信息
	utils.Success(c, menu)
}

// DeleteMenu 删除菜单接口（根据ID删除）
// 路由映射：DELETE /api/v1/menu/:id
// 功能：从URL路径获取菜单ID，调用服务层执行删除操作，返回删除结果
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除菜单失败
//	200：删除成功，返回空数据
func (m *MenuHandler) DeleteMenu(c *gin.Context) {
	// 通过URL地址获取所删除菜单ID的INT格式
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	//调用service层删除菜单
	if err := m.MenuService.DeleteMenu(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 删除成功
	utils.Success(c, nil)
}
