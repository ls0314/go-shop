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
// 接收值：menuService - 菜单服务层对象指针
// 返回值：*MenuHandler - 菜单handler指针
func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

// CreateMenu 创建菜单接口
// 接收值：c - 前端传入JSON参数
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
// 接收值：c - 前端传入JSON参数
func (m *MenuHandler) GetMenu(c *gin.Context) {
	// 通过传入URL地址获取INT格式的ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
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

// GetMenuList 分页查询菜单信息接口
// 接收值：前端传入JSON参数
// @page     查询页数
// @pageSize 查询页面大小
// @menuType 查询菜单类型
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

// GetMenuTree 角色菜单树接口
// 接收值：前端传入JSON参数
// TODO： 构建角色菜单树
//func (m *MenuHandler) GetMenuTree(c *gin.Context) {
//	// 调用service层构建角色菜单树
//	menuTree, err := m.MenuService.GetMenuTree()
//	if err != nil {
//		utils.Error(c, 500, err.Error())
//	}
//	utils.Success(c, menuTree)
//}

// UpdateMenu 更新菜单接口
// 接收值：c - 前端传入JSON参数
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
// 接收值：c - 前端传入JSON参数
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
