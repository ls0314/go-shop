package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MenuPermissionHandler 菜单权限关联表handler层实例
type MenuPermissionHandler struct {
	MenuPermissionService *service.MenuPermissionService // 菜单权限关联服务层对象指针
}

// NewMenuPermissionHandler 新建菜单权限关联表的HTTP handler实例
// 接收值：无接收值
// 返回值：*MenuPermissionHandler - 菜单权限关联handler指针
func NewMenuPermissionHandler() *MenuPermissionHandler {
	return &MenuPermissionHandler{
		MenuPermissionService: service.NewMenuPermissionService(),
	}
}

// CreateMenuPermissionRel 为菜单分配权限接口
// 路由映射：POST /api/v1/menu/assign-perm
// 功能：接收前端传递的菜单ID和权限ID列表，校验参数后调用服务层创建菜单与权限的关联关系
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	menu_id        - 菜单ID，int64类型
//	permission_ids - 权限ID列表，[]int64类型
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建关联失败，返回服务器异常信息
//	200：创建成功，返回传入的菜单ID和权限ID列表
func (mp *MenuPermissionHandler) CreateMenuPermissionRel(c *gin.Context) {
	var menuPermissionIds struct {
		MenuId        int64   `json:"menu_id"`
		PermissionIds []int64 `json:"permission_ids"`
	}
	if err := c.ShouldBind(&menuPermissionIds); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	err := mp.MenuPermissionService.CreateMenuPermission(menuPermissionIds.MenuId, menuPermissionIds.PermissionIds)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, menuPermissionIds)
}

// GetMenuPermissionRelList 根据菜单ID查询关联的权限列表接口
// 路由映射：GET /api/v1/menu/:id/perm
// 功能：从URL路径中获取菜单ID，查询并返回该菜单关联的权限列表及总条数
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询关联权限列表失败
//	200：查询成功，返回关联权限列表和总条数
func (mp *MenuPermissionHandler) GetMenuPermissionRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	menuPermissionList, total, err := mp.MenuPermissionService.GetMenuPermissionList(id)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"List":  menuPermissionList,
		"total": total,
	})
}

// DeleteMenuAllPermissionRel 根据菜单ID清空关联的所有权限接口
// 路由映射：DELETE /api/v1/menu/:id/clear-perm
// 功能：从URL路径获取菜单ID，调用服务层清空该菜单关联的所有权限
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层清空关联权限失败
//	200：清空成功，返回空数据
func (mp *MenuPermissionHandler) DeleteMenuAllPermissionRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	if err = mp.MenuPermissionService.DeleteMenuPermission(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
