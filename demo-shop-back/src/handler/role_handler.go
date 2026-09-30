package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleHandler handler层角色对象实例
type RoleHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewRoleHandler 新建handler层角色对象实例
// 接收值：userRPC - user-service 的 RPC 客户端
// 返回值：*RoleHandler - handler层角色对象指针
func NewRoleHandler(userRPC *userclient.PermCodesClient) *RoleHandler {
	return &RoleHandler{userRPC: userRPC}
}

// CreateRole 创建角色接口
// 路由映射：POST /api/v1/admin/role
func (r *RoleHandler) CreateRole(c *gin.Context) {
	var role model.SysRole
	if err := c.ShouldBind(&role); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	created, errMsg, err := r.userRPC.CreateRole(&role)
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

// GetRole 根据ID获取单个角色信息接口
// 路由映射：GET /api/v1/admin/role/:id
func (r *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	role, errMsg, err := r.userRPC.GetRole(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, role)
}

// GetRoleList 分页获取角色列表接口
// 路由映射：GET /api/v1/admin/role
func (r *RoleHandler) GetRoleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	roleType := c.Query("roleType")

	roleList, total, errMsg, err := r.userRPC.ListRoles(page, pageSize, roleType)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":     roleList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateRole 根据ID更新角色信息接口
// 路由映射：PUT /api/v1/admin/role/:id
func (r *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	// 用 map 接收,只传需要更新的字段
	var updateRole map[string]interface{}
	if err := c.ShouldBind(&updateRole); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	role, errMsg, err := r.userRPC.UpdateRole(id, updateRole)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, role)
}

// DeleteRole 根据ID删除角色接口
// 路由映射：DELETE /api/v1/admin/role/:id
func (r *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := r.userRPC.DeleteRole(id)
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
