package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserRoleHandler 用户角色关联表handler层实例
type UserRoleHandler struct {
	UserRoleService *service.UserRoleService // 用户角色关联服务层对象指针
}

// NewUserRoleHandler 新建用户角色关联表的HTTP handler实例
// 接收值：无接收值
// 返回值：*UserRoleHandler - 用户角色关联handler指针
func NewUserRoleHandler() *UserRoleHandler {
	return &UserRoleHandler{
		UserRoleService: service.NewUserRoleService(),
	}
}

// CreateUserRoleRel 为用户分配角色接口
// 路由映射：POST /api/v1/user/assign-role
// 功能：接收前端传递的用户ID和角色ID列表，校验参数后调用服务层创建用户与角色的关联关系
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	user_id   - 用户ID，int64类型
//	role_ids  - 角色ID列表，[]int64类型
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建关联失败，返回服务器异常信息
//	200：创建成功，返回传入的用户ID和角色ID列表
func (ur *UserRoleHandler) CreateUserRoleRel(c *gin.Context) {
	var userRoleIds struct {
		UserId int64   `json:"user_id"`
		RoleId []int64 `json:"role_ids"`
	}
	if err := c.ShouldBind(&userRoleIds); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	err := ur.UserRoleService.CreateUserRole(userRoleIds.UserId, userRoleIds.RoleId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, userRoleIds)
}

// GetUserRoleRelList 根据用户ID查询关联的角色列表接口
// 路由映射：GET /api/v1/user/:id/role
// 功能：从URL路径中获取用户ID，查询并返回该用户关联的角色列表及总条数
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询关联角色列表失败
//	200：查询成功，返回关联角色列表和总条数
func (ur *UserRoleHandler) GetUserRoleRelList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userRoleList, total, err := ur.UserRoleService.GetUserRoleList(id)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"List":  userRoleList,
		"total": total,
	})
}

// DeleteUserAllRoleRel 根据用户ID清空关联的所有角色接口
// 路由映射：DELETE /api/v1/user/:id/clear-role
// 功能：从URL路径获取用户ID，调用服务层清空该用户关联的所有角色
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层清空关联角色失败
//	200：清空成功，返回空数据
func (ur *UserRoleHandler) DeleteUserAllRoleRel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}
	if err = ur.UserRoleService.DeleteUserRole(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
