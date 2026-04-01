package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	PermService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService}
}

// CreatePermission 创建权限接口
func (p *PermissionHandler) CreatePermission(c *gin.Context) {
	var perm model.SysPermission
	if err := c.ShouldBind(&perm); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	if err := p.PermService.CreatePermission(&perm); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, perm)
}

// GetPermission 获取权限信息接口
func (p *PermissionHandler) GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	perm, err := p.PermService.GetPermission(id)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, perm)
}

// GetPermissionList 分页获取权限信息接口
func (ctrl *PermissionHandler) GetPermissionList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	permType := c.Query("permType") // 可选筛选：permission_type

	perms, total, err := ctrl.PermService.GetPermissionList(page, pageSize, permType)
	if err != nil {
		utils.Error(c, 500, "查询失败")
		return
	}

	// 返回分页数据
	utils.Success(c, gin.H{
		"list":     perms,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdataPermission 更新权限接口
func (p *PermissionHandler) UpdataPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	var perm model.SysPermission
	if err := c.ShouldBind(&perm); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	perm.PermissionID = id
	if err := p.PermService.UpdataPermission(&perm); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, perm)
}

// DeletePermission 删除权限接口
func (p *PermissionHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	if err := p.PermService.DeletePermission(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
