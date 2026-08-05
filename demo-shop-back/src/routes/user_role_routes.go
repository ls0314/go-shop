package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var userRoleCtrl *handler.UserRoleHandler

func InitUserRoleModule() {

	userRoleCtrl = handler.NewUserRoleHandler()
}

func RegisterUserRoleRoutes(r *gin.Engine) {
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	userRoleGroup := r.Group("/api/v1/user")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	userRoleGroup.Use(middleware.AuthMiddleware())
	userRoleGroup.Use(middleware.OperationLogMiddleware(model.LogModuleUser))
	{
		// 批量创建用户角色关联接口
		userRoleGroup.POST("/assign-role", userRoleCtrl.CreateUserRoleRel)
		// 获取某用户全部角色关联列表接口
		userRoleGroup.GET("/:id/role", userRoleCtrl.GetUserRoleRelList)
		// 根据ID删除某用户全部角色关联接口
		userRoleGroup.DELETE("/:id/clear-role", userRoleCtrl.DeleteUserAllRoleRel)
	}
}
