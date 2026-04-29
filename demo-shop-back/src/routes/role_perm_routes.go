package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var rolePermCtrl *handler.RolePermHandler

func InitRolePermModule() {

	rolePermCtrl = handler.NewRolePermHandler()
}

func RegisterRolePermRoutes(r *gin.Engine) {
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	rolePermGroup := r.Group("/api/v1/role")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	rolePermGroup.Use(middleware.AuthMiddleware())
	{
		// 批量创建角色权限关联接口
		rolePermGroup.POST("/assign-perm", rolePermCtrl.CreateRolePermRel)
		// 获取某角色全部权限关联列表接口
		rolePermGroup.GET("/:id/perm", rolePermCtrl.GetRolePermRelList)
		// 根据ID删除某角色全部权限关联接口
		rolePermGroup.DELETE("/:id/clear-perm", rolePermCtrl.DeleteRoleAllPermRelRel)
	}
}
