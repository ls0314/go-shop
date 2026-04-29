package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleMenuHandler 角色菜单关联表handler层实例
type RoleMenuHandler struct {
	RoleMenuService *service.RoleMenuService // 角色菜单关联服务层对象指针
}

// NewRoleMenuHandler 新建角色菜单关联表的HTTP handler实例
// 接收值：无接收值
// 返回值：*RoleMenuHandler - 角色菜单关联handler指针
func NewRoleMenuHandler() *RoleMenuHandler {
	return &RoleMenuHandler{
		RoleMenuService: service.NewRoleMenuService(),
	}
}

// CreateRoleMenuRel 为角色分配菜单接口
// 路由映射：POST /api/v1/role/assign-menu
// 功能：接收前端传递的角色ID和菜单ID列表，校验参数后调用服务层创建角色与菜单的关联关系
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	role_id  - 角色ID，int64类型
//	menu_ids - 菜单ID列表，[]int64类型
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建关联失败，返回服务器异常信息
//	200：创建成功，返回传入的角色ID和菜单ID列表
func (rm *RoleMenuHandler) CreateRoleMenuRel(c *gin.Context) {
	var roleMenuIds struct {
		RoleId  int64   `json:"role_id"`
		MenuIds []int64 `json:"menu_ids"`
	}
	if err := c.ShouldBind(&roleMenuIds); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	if err := rm.RoleMenuService.CreateRoleMenu(roleMenuIds.RoleId, roleMenuIds.MenuIds); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, roleMenuIds)
}

// GetRoleMenuRelList 根据角色ID查询关联的菜单列表接口
// 路由映射：GET /api/v1/role/:id/menu
// 功能：从URL路径中获取角色ID，查询并返回该角色关联的菜单列表及总条数
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询关联菜单列表失败
//	200：查询成功，返回关联菜单列表和总条数
func (rm *RoleMenuHandler) GetRoleMenuRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	roleMenuList, total, err := rm.RoleMenuService.GetRoleMenuList(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"list":  roleMenuList,
		"total": total,
	})
}

// DeleteRoleAllMenuRel 根据角色ID清空关联的所有菜单接口
// 路由映射：DELETE /api/v1/role/:id/clear-menu
// 功能：从URL路径获取角色ID，调用服务层清空该角色关联的所有菜单
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层清空关联菜单失败
//	200：清空成功，返回空数据
func (rm *RoleMenuHandler) DeleteRoleAllMenuRel(c *gin.Context) {
	// 从URL路径参数中获取待删除的权限范围ID字符串
	idStr := c.Param("id")
	// 将字符串ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// ID格式转换失败，返回参数错误响应
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	if err := rm.RoleMenuService.DeleteRoleMenuByRoleId(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
