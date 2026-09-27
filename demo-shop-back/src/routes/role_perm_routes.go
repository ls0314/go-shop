package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var rolePermCtrl *handler.RolePermHandler

func InitRolePermModule(deps service.ServiceDeps) {

	rolePermCtrl = handler.NewRolePermHandler(deps)
}

func RegisterRolePermRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	rolePermGroup := r.Group("/api/v1/role")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	rolePermGroup.Use(middleware.AuthMiddleware())
	rolePermGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		// 批量创建角色权限关联接口
		rolePermGroup.POST("/assign-perm", permMW, rolePermCtrl.CreateRolePermRel)
		// 获取某角色全部权限关联列表接口
		rolePermGroup.GET("/:id/perm", permMW, rolePermCtrl.GetRolePermRelList)
		// 根据ID删除某角色全部权限关联接口
		rolePermGroup.DELETE("/:id/clear-perm", permMW, rolePermCtrl.DeleteRoleAllPermRelRel)
	}
}
