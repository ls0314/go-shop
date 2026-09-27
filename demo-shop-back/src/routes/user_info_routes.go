package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var userInfoCtrl *handler.UserInfoHandler

func InitUserInfoModule(deps service.ServiceDeps) {

	userInfoCtrl = handler.NewUserInfoHandler(deps)
}

func RegisterUserInfoRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 创建角色接口路由分组，统一前缀 /api/v1/Info
	userInfoGroup := r.Group("/api/v1/user/info")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	userInfoGroup.Use(middleware.AuthMiddleware())
	{
		// 创建角色信息接口
		userInfoGroup.POST("create", userInfoCtrl.CreateUserInfo)
		// 根据ID获取单个角色信息接口
		userInfoGroup.GET("/:id", userInfoCtrl.GetUserInfo)
		// 根据ID更新角色信息接口
		userInfoGroup.PUT("/:id", userInfoCtrl.UpdateUserInfo)
		// 根据ID删除角色信息接口
		userInfoGroup.DELETE("/:id", userInfoCtrl.DeleteUserInfo)
	}
}
