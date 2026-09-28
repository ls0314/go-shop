package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserInfoHandler 用户信息表handler层实例
type UserInfoHandler struct {
	UserInfoService *service.UserInfoService // 用户信息服务层对象指针
}

// NewUserInfoHandler 新建用户信息表的HTTP handler实例
// 接收值：无接收值
// 返回值：*UserInfoHandler - 用户信息handler指针
func NewUserInfoHandler(deps service.ServiceDeps) *UserInfoHandler {
	return &UserInfoHandler{
		UserInfoService: service.NewUserInfoService(deps),
	}
}

// CreateUserInfo 创建用户信息接口
// 路由映射：POST /api/v1/user/user-info
// 功能：接收前端传递的用户信息，校验参数后调用服务层创建用户信息并入库
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建用户信息失败，返回服务器异常信息
//	200：创建成功，返回创建完成的用户信息
func (ui *UserInfoHandler) CreateUserInfo(c *gin.Context) {
	var userInfo *model.UserProfile
	if err := c.ShouldBind(&userInfo); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	userInfo.UserId = userId
	if err := ui.UserInfoService.CreateUserInfo(userInfo); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, userInfo)
}

// GetUserInfo 根据用户ID获取用户信息接口
// 路由映射：GET /api/v1/user/user-info/:id
// 功能：从URL路径中获取用户ID，查询并返回对应用户信息详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询用户信息失败
//	200：查询成功，返回用户信息详细信息
func (ui *UserInfoHandler) GetUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if userId != id {
		utils.Fail(c, 400, model.UserInfoError.Error())
		c.Abort()
		return
	}

	userInfo, err := ui.UserInfoService.GetUserInfoByUserId(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, userInfo)
}

// UpdateUserInfo 根据用户ID更新用户信息接口
// 路由映射：PUT /api/v1/user/user-info/:id
// 功能：从URL获取用户ID，接收前端传入的更新字段，执行用户信息更新，返回更新后的用户信息详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id             - URL路径参数，用户ID
//	updateUserInfo - 请求体JSON，需要更新的用户信息字段（map格式）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新用户信息失败 / 查询更新后用户信息失败
//	200：更新成功，返回更新后的完整用户信息
func (ui *UserInfoHandler) UpdateUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if userId != id {
		utils.Fail(c, 400, model.UserInfoError.Error())
		c.Abort()
		return
	}

	var updateUserInfo map[string]interface{}
	if err := c.ShouldBind(&updateUserInfo); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := ui.UserInfoService.UpdateUserProfile(id, updateUserInfo); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	userInfo, err := ui.UserInfoService.GetUserInfoByUserId(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, userInfo)

}

// DeleteUserInfo 根据用户ID删除用户信息接口
// 路由映射：DELETE /api/v1/user/user-info/:id
// 功能：从URL路径获取用户ID，调用服务层执行删除操作，返回删除结果
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除用户信息失败
//	200：删除成功，返回空数据
func (ui *UserInfoHandler) DeleteUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if userId != id {
		utils.Fail(c, 400, model.UserInfoError.Error())
		c.Abort()
		return
	}
	if err = ui.UserInfoService.DeleteUserInfo(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
