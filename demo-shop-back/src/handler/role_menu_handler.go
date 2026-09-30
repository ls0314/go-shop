package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleMenuHandler 角色菜单关联表handler层实例
type RoleMenuHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewRoleMenuHandler 新建角色菜单关联表的HTTP handler实例
func NewRoleMenuHandler(userRPC *userclient.PermCodesClient) *RoleMenuHandler {
	return &RoleMenuHandler{userRPC: userRPC}
}

// CreateRoleMenuRel 为角色分配菜单接口(全量替换)
// 路由映射：POST /api/v1/admin/role/assign-menu
func (rm *RoleMenuHandler) CreateRoleMenuRel(c *gin.Context) {
	var req struct {
		RoleId  int64   `json:"role_id"`
		MenuIds []int64 `json:"menu_ids"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := rm.userRPC.AssignRoleMenus(req.RoleId, req.MenuIds)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, req)
}

// GetRoleMenuRelList 根据角色ID查询关联的菜单列表接口
// 路由映射：GET /api/v1/admin/role/:id/menu
func (rm *RoleMenuHandler) GetRoleMenuRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	roleMenuList, total, errMsg, err := rm.userRPC.ListRoleMenus(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":  roleMenuList,
		"total": total,
	})
}

// DeleteRoleAllMenuRel 根据角色ID清空关联的所有菜单接口
// 路由映射：DELETE /api/v1/admin/role/:id/clear-menu
func (rm *RoleMenuHandler) DeleteRoleAllMenuRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := rm.userRPC.ClearRoleMenus(id)
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
