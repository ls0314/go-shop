package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserRoleHandler 用户角色关联表handler层实例
type UserRoleHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewUserRoleHandler 新建用户角色关联表的HTTP handler实例
func NewUserRoleHandler(userRPC *userclient.PermCodesClient) *UserRoleHandler {
	return &UserRoleHandler{userRPC: userRPC}
}

// CreateUserRoleRel 为用户分配角色接口(全量替换)
// 路由映射：POST /api/v1/admin/user/assign-role
func (ur *UserRoleHandler) CreateUserRoleRel(c *gin.Context) {
	var req struct {
		UserId  int64   `json:"user_id"`
		RoleIds []int64 `json:"role_ids"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	errMsg, err := ur.userRPC.AssignUserRoles(req.UserId, req.RoleIds)
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

// GetUserRoleRelList 根据用户ID查询关联的角色列表接口
// 路由映射：GET /api/v1/admin/user/:id/role
func (ur *UserRoleHandler) GetUserRoleRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userRoleList, total, errMsg, err := ur.userRPC.ListUserRoles(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":  userRoleList,
		"total": total,
	})
}

// DeleteUserAllRoleRel 根据用户ID清空关联的所有角色接口
// 路由映射：DELETE /api/v1/admin/user/:id/clear-role
func (ur *UserRoleHandler) DeleteUserAllRoleRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := ur.userRPC.ClearUserRoles(id)
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
