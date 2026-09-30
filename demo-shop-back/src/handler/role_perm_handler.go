package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RolePermHandler 角色权限关联表handler层实例
type RolePermHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewRolePermHandler 新建角色权限关联表的HTTP handler实例
func NewRolePermHandler(userRPC *userclient.PermCodesClient) *RolePermHandler {
	return &RolePermHandler{userRPC: userRPC}
}

// CreateRolePermRel 为角色分配权限接口(全量替换)
// 路由映射：POST /api/v1/admin/role/assign-perm
func (rp *RolePermHandler) CreateRolePermRel(c *gin.Context) {
	var req struct {
		RoleId  int64   `json:"role_id"`
		PermIds []int64 `json:"perm_ids"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := rp.userRPC.AssignRolePerms(req.RoleId, req.PermIds)
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

// GetRolePermRelList 根据角色ID查询关联的权限列表接口
// 路由映射：GET /api/v1/admin/role/:id/perm
func (rp *RolePermHandler) GetRolePermRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	rolePermList, total, errMsg, err := rp.userRPC.ListRolePerms(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":  rolePermList,
		"total": total,
	})
}

// DeleteRoleAllPermRelRel 根据角色ID清空关联的所有权限接口
// 路由映射：DELETE /api/v1/admin/role/:id/clear-perm
func (rp *RolePermHandler) DeleteRoleAllPermRelRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := rp.userRPC.ClearRolePerms(id)
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
