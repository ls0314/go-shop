package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserInfoHandler 用户信息表handler层实例
type UserInfoHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewUserInfoHandler 新建用户信息表的HTTP handler实例
func NewUserInfoHandler(userRPC *userclient.PermCodesClient) *UserInfoHandler {
	return &UserInfoHandler{
		userRPC: userRPC,
	}
}

// CreateUserInfo 创建当前登录用户的档案
// 路由映射：POST /api/v1/admin/user/info/create
func (ui *UserInfoHandler) CreateUserInfo(c *gin.Context) {
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	var profile model.UserProfile
	if err := c.ShouldBind(&profile); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 归属强制为当前登录用户,不接受请求体指定
	profile.UserId = userId

	created, errMsg, err := ui.userRPC.CreateUserProfile(&profile)
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

// GetUserInfo 获取当前登录用户的档案
// 路由映射：GET /api/v1/admin/user/info/:id
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

	profile, errMsg, err := ui.userRPC.GetUserProfile(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, profile)
}

// UpdateUserInfo 更新当前登录用户的档案
// 路由映射：PUT /api/v1/admin/user/info/:id
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

	profile, errMsg, err := ui.userRPC.UpdateUserProfile(id, updateUserInfo)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, profile)
}

// DeleteUserInfo 删除当前登录用户的档案
// 路由映射：DELETE /api/v1/admin/user/info/:id
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

	errMsg, err := ui.userRPC.DeleteUserProfile(id)
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
