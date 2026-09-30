package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限表handler层实例
type PermissionHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewPermissionHandler 构建权限表中的HTTP中handler实例
// 接收值：permService - 权限服务层对象指针
// 返回值：*PermissionHandler - 权限handler指针
func NewPermissionHandler(userRPC *userclient.PermCodesClient) *PermissionHandler {
	return &PermissionHandler{userRPC: userRPC}
}

// CreatePermission 创建权限接口、
// 接收值：c - 前端传入JSON参数
func (p *PermissionHandler) CreatePermission(c *gin.Context) {
	// 实例化后绑定参数
	var perm model.SysPermission
	if err := c.ShouldBind(&perm); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 调用服务层创建菜单
	_, errMsg, err := p.userRPC.CreatePermission(&perm)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	// 创建成功，返回新建菜单数据
	utils.Success(c, perm)
}

// GetPermission 查询权限信息接口（根据ID查询）
// 接收值：c - 前端传入JSON参数
func (p *PermissionHandler) GetPermission(c *gin.Context) {
	// 通过传入URL地址获取INT格式的ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	// 调用服务层查找权限信息
	perm, errMsg, err := p.userRPC.GetPermission(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	// 查询成功，返回所查询权限信息
	utils.Success(c, perm)
}

// GetPermissionList 分页查询权限信息接口
// 接收值：前端传入JSON参数
// @page     查询页数
// @pageSize 查询页面大小
// @permType 查询权限类型
func (p *PermissionHandler) GetPermissionList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	permType := c.Query("permType")
	// 调用服务层分页查询权限
	perms, total, errMsg, err := p.userRPC.ListPermissions(page, pageSize, permType)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
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

// UpdatePermission 更新权限接口
// 接收值：c - 前端传入JSON参数
func (p *PermissionHandler) UpdatePermission(c *gin.Context) {
	// 通过URL地址获取所更新权限ID的INT格式（json传进格式一般为float64，故此处从URL取id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 实例化后绑定参数，此处使用map来确保权限信息可以局部更新
	var updatePerm map[string]interface{}
	if err := c.ShouldBind(&updatePerm); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	//调用服务层更新权限部分信息
	_, errMsg, err := p.userRPC.UpdatePermission(id, updatePerm)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	// 获取权限更新后的完整信息
	perm, errMsg, err := p.userRPC.GetPermission(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	// 更新成功,返回更新后完整权限信息
	utils.Success(c, perm)
}

// DeletePermission 删除权限接口（根据ID删除）
// 接收值：c - 前端传入JSON参数
func (p *PermissionHandler) DeletePermission(c *gin.Context) {
	// 通过URL地址获取所删除权限ID的INT格式
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	//调用service层删除权限
	errMsg, err := p.userRPC.DeletePermission(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	// 删除成功
	utils.Success(c, nil)
}
