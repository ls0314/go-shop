package handler

import (
	"demo-shop-back/src/infra/userclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户表handler层实例
type UserHandler struct {
	userRPC *userclient.PermCodesClient
}

// NewUserHandler 新建用户表的HTTP handler实例
func NewUserHandler(userRPC *userclient.PermCodesClient) *UserHandler {
	return &UserHandler{
		userRPC: userRPC,
	}
}

// GetUserInfoByContext 从JWT鉴权上下文中提取当前登录用户ID
// 接收值：c - Gin上下文
// 返回值：int64 - 用户ID,userName - 用户名, error - 用户未登录时返回UserNotLogin
func GetUserInfoByContext(c *gin.Context) (int64, string, error) {
	// 从上下文中获取中间件注入的user_id
	userIdVal, exist := c.Get("user_id")
	if !exist {
		return 0, "", model.UserNotLogin
	}
	userNameVal, exist := c.Get("username")
	if !exist {
		return 0, "", model.UserNotLogin
	}
	// 类型断言确保为int64
	userId, ok := userIdVal.(int64)
	if !ok {
		return 0, "", model.UserNotLogin
	}
	userName, ok := userNameVal.(string)
	if !ok {
		return 0, "", model.UserNotLogin
	}
	return userId, userName, nil
}

// CreateUserHandler 创建用户接口(管理端新增用户)
// 路由映射：POST /api/v1/admin/user
func (u *UserHandler) CreateUserHandler(c *gin.Context) {
	var req model.SysUser
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, errMsg, err := u.userRPC.CreateUser(&req, req.PasswordHash)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, user)
}

// RegisterHandler 自助注册接口
// 路由映射：POST /api/v1/user/register
// 与 CreateUserHandler(管理端建号)的区别：注册会校验手机号格式、查重用户名/手机号/邮箱,
// 并在同一事务内创建 user_profile(nickname 为空时回落为用户名)。
func (u *UserHandler) RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, errMsg, err := u.userRPC.Register(req.Username, req.Password, req.Phone, req.Email, req.Nickname)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		// 口令/手机号格式不合法、用户名或手机号已存在等均属业务失败,用 400
		utils.Fail(c, 400, errMsg)
		return
	}

	utils.Success(c, user)
}

// GetUserInfo 获取当前登录用户信息接口
// 路由映射：GET /api/v1/user/info
func (u *UserHandler) GetUserInfo(c *gin.Context) {
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.UserNotLogin.Error())
		return
	}

	user, profile, errMsg, err := u.userRPC.GetSelfInfo(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"user_id":  user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"phone":    user.Phone,
		"status":   user.Status,
		"profile":  profile,
	})
}

// GetUserPerms 获取当前登录用户的全部权限码接口
// 路由映射：GET /api/v1/user/perms
func (u *UserHandler) GetUserPerms(c *gin.Context) {
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.UserNotLogin.Error())
		return
	}

	codes, errMsg, err := u.userRPC.ListSelfPermCodes(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	if codes == nil {
		codes = []string{}
	}
	utils.Success(c, gin.H{"perms": codes})
}

// GetUser 根据ID获取用户接口
// 路由映射：GET /api/v1/admin/user/:id
func (u *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, errMsg, err := u.userRPC.GetUser(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, user)
}

// GetUserList 分页获取用户列表接口
// 路由映射：GET /api/v1/admin/user
func (u *UserHandler) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	users, total, errMsg, err := u.userRPC.ListUsers(page, pageSize, status)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, gin.H{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateUser 根据ID更新用户接口
// 路由映射：PUT /api/v1/admin/user/:id
func (u *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	var updateUser map[string]interface{}
	if err := c.ShouldBind(&updateUser); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, errMsg, err := u.userRPC.UpdateUser(id, updateUser)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}

	utils.Success(c, user)
}

// DeleteUser 根据ID删除用户接口
// 路由映射：DELETE /api/v1/admin/user/:id
func (u *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	errMsg, err := u.userRPC.DeleteUser(id)
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

// LoginHandler 用户登录接口
// 路由映射：POST /api/v1/user/login
func (u *UserHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	resp, err := u.userRPC.Login(req.Username, req.Phone, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if resp.ErrorMsg != "" {
		// 凭据错误用 401,前端 axios 拦截器据此判断是否需要重新登录
		c.JSON(401, gin.H{"code": 401, "message": resp.ErrorMsg, "data": nil})
		return
	}

	utils.Success(c, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// RefreshHandler 刷新访问令牌接口
// 路由映射：POST /api/v1/user/refresh
func (u *UserHandler) RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := u.userRPC.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if resp.ErrorMsg != "" {
		utils.Fail(c, 400, resp.ErrorMsg)
		return
	}

	utils.Success(c, gin.H{"access_token": resp.AccessToken})
}
