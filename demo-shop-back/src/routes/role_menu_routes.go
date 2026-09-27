package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var roleMenuCtrl *handler.RoleMenuHandler

func InitRoleMenuModule(deps service.ServiceDeps) {
	roleMenuCtrl = handler.NewRoleMenuHandler(deps)
}

func RegisterRoleMenuRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	roleMenuGroup := r.Group("/api/v1/role")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	roleMenuGroup.Use(middleware.AuthMiddleware())
	roleMenuGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		// 批量创建角色菜单关联接口
		roleMenuGroup.POST("/assign-menu", permMW, roleMenuCtrl.CreateRoleMenuRel)
		// 获取某角色全部菜单关联列表接口
		roleMenuGroup.GET("/:id/menu", permMW, roleMenuCtrl.GetRoleMenuRelList)
		// 根据ID删除某角色全部菜单关联接口
		roleMenuGroup.DELETE("/:id/clear-menu", permMW, roleMenuCtrl.DeleteRoleAllMenuRel)
	}
}
