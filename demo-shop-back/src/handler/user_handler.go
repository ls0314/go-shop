package handler

import (
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户表handler层实例
type UserHandler struct {
	UserService *service.UserService // 用户服务层对象指针
	permRepo    *repository.PermissionRepo
	cache       *cache.RedisService
}

// NewUserHandler 新建用户表的HTTP handler实例
// 接收值：无接收值
// 返回值：*UserHandler - 用户handler指针
func NewUserHandler(deps service.ServiceDeps) *UserHandler {
	return &UserHandler{
		UserService: service.NewUserService(deps),
		permRepo:    repository.NewPermissionRepo(deps.DB),
		cache:       deps.Cache,
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

// CreateUserHandler 用户注册接口
// 路由映射：POST /api/v1/user/register
// 功能：接收前端传递的用户注册信息，校验参数后调用服务层执行注册
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	req - 请求体JSON，用户注册信息（model.SysUser类型）
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层注册失败，返回服务器异常信息
//	200：注册成功，返回注册的用户信息
func (u *UserHandler) CreateUserHandler(c *gin.Context) {

	var req model.SysUser

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	err := u.UserService.CreateUser(&req)

	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, req)

}

// GetUserInfo 获取当前登录用户信息接口
// 路由映射：GET /api/v1/user/info
// 功能：从Gin上下文中获取当前登录用户的ID和用户名并返回
// 参数：c *gin.Context Gin上下文，用于获取上下文数据、返回响应
// 响应：
//
//	200：查询成功，返回用户ID和用户名
//
// GetUserInfo 获取当前登录用户信息接口
// 路由映射：GET /api/v1/user/info
// 功能：从 JWT 上下文取出 user_id/username 返回(前端登录后展示用)
func (u *UserHandler) GetUserInfo(c *gin.Context) {

	userID, _ := c.Get("user_id") //
	username, _ := c.Get("username")

	utils.Success(c, gin.H{"user_id": userID,
		"username": username})
}

// GetUserPerms 获取当前登录用户的全部权限码接口
// 路由映射：GET /api/v1/user/perms
// 鉴权：AuthMiddleware(仅需登录)
// 功能：返回用户权限码数组(如 ["system:user:view","system:role:create",...])，
//
//	前端用于按钮级权限控制(有权限才渲染新增/编辑/删除按钮)
func (u *UserHandler) GetUserPerms(c *gin.Context) {
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.UserNotLogin.Error())
		return
	}
	codes, err := middleware.GetUserPermCodes(u.permRepo, u.cache, userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"perms": codes})
}

func (u *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, err := u.UserService.GetUser(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, user)
}

func (u *UserHandler) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	userList, total, err := u.UserService.GetUserList(page, pageSize, status)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":     userList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (u *UserHandler) UpdateUser(c *gin.Context) {
	// 通过URL地址获取所更新用户ID的INT格式（json传进格式一般为float64，故此处从URL取id）
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	// 实例化后绑定参数, 此处使用map来确保用户信息可以局部更新
	var updateUser map[string]interface{}
	if err := c.ShouldBind(&updateUser); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}
	//调用服务层更新用户部分信息
	if err := u.UserService.UpdateUser(id, updateUser); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 获取用户更新后的完整信息
	user, err := u.UserService.GetUser(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
	}
	// 更新成功,返回更新后完整用户信息
	utils.Success(c, user)
}

func (u *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	user, err := u.UserService.GetUser(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	if err := u.UserService.DeleteUser(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, user)
}

// LoginHandler 用户登录接口
// 路由映射：POST /api/v1/user/login
// 功能：接收前端传递的登录信息，获取客户端IP和设备信息，调用服务层执行登录并返回Token
// 参数：c *gin.Context Gin上下文，用于接收请求参数、获取客户端信息、返回响应
// 请求参数：
//
//	req - 请求体JSON，登录信息（model.LoginRequest类型）
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层登录失败，返回服务器异常信息
//	200：登录成功，返回登录响应（含Token等）
func (u *UserHandler) LoginHandler(c *gin.Context) {

	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	ip := c.ClientIP()
	device := c.Request.UserAgent()

	resp, err := u.UserService.Login(&req, ip, device)

	if err != nil {

		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// RefreshHandler 刷新访问令牌接口
// 路由映射：POST /api/v1/user/refresh
// 功能：接收前端传递的刷新令牌，校验后生成新的访问令牌并返回
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	refresh_token - 刷新令牌，string类型
//
// 响应：
//
//	400：请求参数绑定失败/RefreshToken无效或已过期/Token类型错误/生成Token失败，返回错误信息
//	200：刷新成功，返回新的访问令牌
func (u *UserHandler) RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	jwtService := middleware.GetJWTService()

	claims, err := jwtService.ParseToken(req.RefreshToken)
	if err != nil {
		utils.Fail(c, 400, "RefreshToken无效/已过期")
		return
	}

	if claims.TokenType != "refresh" {
		utils.Fail(c, 400, "Token类型错误")
		return
	}

	newAccessToken, err := jwtService.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		utils.Fail(c, 400, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"access_token": newAccessToken,
	})
}
