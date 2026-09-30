package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ScopeHandler handler层权限范围对象实例
type ScopeHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewScopeHandler 新建handler层权限范围对象实例
func NewScopeHandler(userRPC *userclient.PermCodesClient) *ScopeHandler {
	return &ScopeHandler{userRPC: userRPC}
}

// CreateScope 创建权限范围接口
// 路由映射：POST /api/v1/admin/scope
func (s *ScopeHandler) CreateScope(c *gin.Context) {
	var scope model.SysScope
	if err := c.ShouldBind(&scope); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	created, errMsg, err := s.userRPC.CreateScope(&scope)
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

// GetScope 根据ID获取单个权限范围信息接口
// 路由映射：GET /api/v1/admin/scope/:id
func (s *ScopeHandler) GetScope(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist)
		return
	}

	scope, errMsg, err := s.userRPC.GetScope(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, scope)
}

// GetScopeList 分页获取权限范围列表接口
// 路由映射：GET /api/v1/admin/scope
func (s *ScopeHandler) GetScopeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	resourceType := c.Query("resourceType")

	scopes, total, errMsg, err := s.userRPC.ListScopes(page, pageSize, resourceType)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":     scopes,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateScope 根据ID更新权限范围信息接口
// 路由映射：PUT /api/v1/admin/scope/:id
func (s *ScopeHandler) UpdateScope(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	// 用 map 接收,只传需要更新的字段
	var updateScope map[string]interface{}
	if err := c.ShouldBind(&updateScope); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	scope, errMsg, err := s.userRPC.UpdateScope(id, updateScope)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, scope)
}

// DeleteScope 根据ID删除权限范围接口
// 路由映射：DELETE /api/v1/admin/scope/:id
func (s *ScopeHandler) DeleteScope(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	errMsg, err := s.userRPC.DeleteScope(id)
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
