package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"

	"github.com/gin-gonic/gin"
)

var roleMenuCtrl *handler.RoleMenuHandler

func InitRoleMenuModule() {
	roleMenuCtrl = handler.NewRoleMenuHandler()
}

func RegisterRoleMenuRoutes(r *gin.Engine) {
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	roleMenuGroup := r.Group("/api/v1/role")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	roleMenuGroup.Use(middleware.AuthMiddleware())
	roleMenuGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		// 批量创建角色菜单关联接口
		roleMenuGroup.POST("/assign-menu", roleMenuCtrl.CreateRoleMenuRel)
		// 获取某角色全部菜单关联列表接口
		roleMenuGroup.GET("/:id/menu", roleMenuCtrl.GetRoleMenuRelList)
		// 根据ID删除某角色全部菜单关联接口
		roleMenuGroup.DELETE("/:id/clear-menu", roleMenuCtrl.DeleteRoleAllMenuRel)
	}
}
