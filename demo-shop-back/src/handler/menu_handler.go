package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单表handler层实例
type MenuHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewMenuHandler 新建菜单表中的HTTP handler实例
func NewMenuHandler(userRPC *userclient.PermCodesClient) *MenuHandler {
	return &MenuHandler{userRPC: userRPC}
}

// CreateMenu 创建菜单接口
// 路由映射：POST /api/v1/admin/menu
func (m *MenuHandler) CreateMenu(c *gin.Context) {
	var menu model.SysMenu
	if err := c.ShouldBind(&menu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	created, errMsg, err := m.userRPC.CreateMenu(&menu)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, created)
}

// GetMenu 查询菜单信息接口（根据ID查询）
// 路由映射：GET /api/v1/admin/menu/:id
func (m *MenuHandler) GetMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	menu, errMsg, err := m.userRPC.GetMenu(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, menu)
}

// GetMenuTreeByRoleId 根据角色ID获取对应的菜单树
// 路由映射：GET /api/v1/admin/menu/:id/tree
func (m *MenuHandler) GetMenuTreeByRoleId(c *gin.Context) {
	roleIdStr := c.Param("id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	menuTree, err := m.userRPC.GetMenuTreeByRoleId(roleId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, menuTree)
}

// GetMenuTreeByUserId 查询当前登录用户的菜单树
// 路由映射：POST /api/v1/admin/menu/tree
// 用户ID取自 JWT,不接受请求体指定:否则任何登录用户都能查到他人的菜单结构。
func (m *MenuHandler) GetMenuTreeByUserId(c *gin.Context) {
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.UserNotLogin.Error())
		return
	}

	menuTree, errMsg, err := m.userRPC.GetMenuTreeByUserId(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, menuTree)
}

// GetMenuList 分页查询菜单信息接口
// 路由映射：GET /api/v1/admin/menu
func (m *MenuHandler) GetMenuList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	menuType := c.Query("menuType")

	menu, total, errMsg, err := m.userRPC.ListMenus(page, pageSize, menuType)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":     menu,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateMenu 更新菜单接口
// 路由映射：PUT /api/v1/admin/menu/:id
func (m *MenuHandler) UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	// 用 map 接收,只传需要更新的字段
	var updateMenu map[string]interface{}
	if err := c.ShouldBind(&updateMenu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	menu, errMsg, err := m.userRPC.UpdateMenu(id, updateMenu)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, menu)
}

// DeleteMenu 删除菜单接口（根据ID删除）
// 路由映射：DELETE /api/v1/admin/menu/:id
func (m *MenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := m.userRPC.DeleteMenu(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, nil)
}
