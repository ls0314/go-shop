package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var userRoleCtrl *handler.UserRoleHandler

func InitUserRoleModule(deps service.ServiceDeps) {

	userRoleCtrl = handler.NewUserRoleHandler(deps.UserRPC)
}

func RegisterUserRoleRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(deps.UserRPC, deps.Cache)
	// 管理端:用户-角色关联。前缀统一 /api/v1/admin/**
	userRoleGroup := r.Group("/api/v1/admin/user")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	userRoleGroup.Use(middleware.AuthMiddleware())
	userRoleGroup.Use(middleware.OperationLogMiddleware(model.LogModuleUser))
	{
		// 批量创建用户角色关联接口
		userRoleGroup.POST("/assign-role", permMW, userRoleCtrl.CreateUserRoleRel)
		// 获取某用户全部角色关联列表接口
		userRoleGroup.GET("/:id/role", permMW, userRoleCtrl.GetUserRoleRelList)
		// 根据ID删除某用户全部角色关联接口
		userRoleGroup.DELETE("/:id/clear-role", permMW, userRoleCtrl.DeleteUserAllRoleRel)
	}
}
