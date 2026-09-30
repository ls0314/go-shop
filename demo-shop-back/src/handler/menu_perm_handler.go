package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MenuPermissionHandler 菜单权限关联表handler层实例
type MenuPermissionHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewMenuPermissionHandler 新建菜单权限关联表的HTTP handler实例
func NewMenuPermissionHandler(userRPC *userclient.PermCodesClient) *MenuPermissionHandler {
	return &MenuPermissionHandler{userRPC: userRPC}
}

// CreateMenuPermissionRel 为菜单分配权限接口(全量替换)
// 路由映射：POST /api/v1/admin/menu/assign-perm
func (mp *MenuPermissionHandler) CreateMenuPermissionRel(c *gin.Context) {
	var req struct {
		MenuId        int64   `json:"menu_id"`
		PermissionIds []int64 `json:"permission_ids"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := mp.userRPC.AssignMenuPerms(req.MenuId, req.PermissionIds)
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

// GetMenuPermissionRelList 根据菜单ID查询关联的权限列表接口
// 路由映射：GET /api/v1/admin/menu/:id/perm
func (mp *MenuPermissionHandler) GetMenuPermissionRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	menuPermissionList, total, errMsg, err := mp.userRPC.ListMenuPerms(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":  menuPermissionList,
		"total": total,
	})
}

// DeleteMenuAllPermissionRel 根据菜单ID清空关联的所有权限接口
// 路由映射：DELETE /api/v1/admin/menu/:id/clear-perm
func (mp *MenuPermissionHandler) DeleteMenuAllPermissionRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := mp.userRPC.ClearMenuPerms(id)
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
