package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RolePermHandler 角色权限关联表handler层实例
type RolePermHandler struct {
	RolePermService *service.RolePermService // 角色权限关联服务层对象指针
}

// NewRolePermHandler 新建角色权限关联表的HTTP handler实例
// 接收值：无接收值
// 返回值：*RolePermHandler - 角色权限关联handler指针
func NewRolePermHandler(deps service.ServiceDeps) *RolePermHandler {
	return &RolePermHandler{
		RolePermService: service.NewRolePermService(deps),
	}
}

// CreateRolePermRel 为角色分配权限接口
// 路由映射：POST /api/v1/role/assign-perm
// 功能：接收前端传递的角色ID和权限ID列表，校验参数后调用服务层创建角色与权限的关联关系
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	role_id  - 角色ID，int64类型
//	perm_ids - 权限ID列表，[]int64类型
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建关联失败，返回服务器异常信息
//	200：创建成功，返回传入的角色ID和权限ID列表
func (rp *RolePermHandler) CreateRolePermRel(c *gin.Context) {
	var rolePermIds struct {
		RoleId  int64   `json:"role_id"`
		PermIds []int64 `json:"perm_ids"`
	}
	if err := c.ShouldBind(&rp); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	if err := rp.RolePermService.CreateRolePerm(rolePermIds.RoleId, rolePermIds.PermIds); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, rolePermIds)
}

// GetRolePermRelList 根据角色ID查询关联的权限列表接口
// 路由映射：GET /api/v1/role/:id/perm
// 功能：从URL路径中获取角色ID，查询并返回该角色关联的权限列表及总条数
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询关联权限列表失败
//	200：查询成功，返回关联权限列表和总条数
func (rp *RolePermHandler) GetRolePermRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 500, err.Error())
	}

	rolePermList, total, err := rp.RolePermService.GetRolePermList(id)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"list":  rolePermList,
		"total": total,
	})
}

// DeleteRoleAllPermRelRel 根据角色ID清空关联的所有权限接口
// 路由映射：DELETE /api/v1/role/:id/clear-perm
// 功能：从URL路径获取角色ID，调用服务层清空该角色关联的所有权限
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层清空关联权限失败
//	200：清空成功，返回空数据
func (rp *RolePermHandler) DeleteRoleAllPermRelRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := rp.RolePermService.DeleteRolePermRel(id); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
