package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	MenuService *service.MenuService
}

func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: menuService}
}

func (m *MenuHandler) CreateMenu(c *gin.Context) {
	var menu model.SysMenu
	if err := c.ShouldBind(&menu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	if err := m.MenuService.CreateMenu(&menu); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, menu)
}

func (m *MenuHandler) GetMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
	}
	menu, err := m.MenuService.GetMenu(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, menu)
}

func (ctrl *MenuHandler) GetMenuList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	permType := c.Query("menu_type") // 可选筛选：permission_type

	menu, total, err := ctrl.MenuService.GetMenuList(page, pageSize, permType)
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

func (m *MenuHandler) GetMenuTree(c *gin.Context) {
	menuTree, err := m.MenuService.GetMenuTree()
	if err != nil {
		utils.Error(c, 500, err.Error())
	}
	utils.Success(c, menuTree)
}

func (m *MenuHandler) UpdataMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	var menu model.SysMenu
	if err := c.ShouldBind(&menu); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	menu.MenuId = id
	if err := m.MenuService.UpdateMenu(&menu); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, menu)
}

// DeletePermission 删除权限接口
func (m *MenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	if err := m.MenuService.DeleteMenu(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
