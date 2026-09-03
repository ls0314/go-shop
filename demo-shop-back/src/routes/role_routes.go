package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

// roleCtrl 角色模块全局控制器实例
// 作用：供路由注册时使用，持有角色Handler单例对象
var roleCtrl *handler.RoleHandler

// InitRoleModule 初始化角色模块
// 功能：完成角色模块 仓库层 → 服务层 → 控制层 的依赖注入与实例化
// 执行顺序：创建数据访问层实例 → 创建业务逻辑层实例 → 创建控制器实例
func InitRoleModule() {
	// 初始化角色服务层
	roleService := service.NewRoleService()
	// 初始化角色控制器，赋值给全局控制器变量
	roleCtrl = handler.NewRoleHandler(roleService)
}

// RegisterRoleRoutes 注册角色模块路由
// 参数：r *gin.Engine Gin路由引擎实例
// 功能：注册角色相关API路由，统一前缀 /api/v1/role，并添加登录认证中间件
func RegisterRoleRoutes(r *gin.Engine) {
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	roleGroup := r.Group("/api/v1/role")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	roleGroup.Use(middleware.AuthMiddleware())
	roleGroup.Use(middleware.OperationLogMiddleware(model.LogModulePermission))
	{
		// 创建角色接口
		roleGroup.POST("", middleware.PermissionMiddleware(), roleCtrl.CreateRole)
		// 分页获取角色列表接口
		roleGroup.GET("", middleware.PermissionMiddleware(), roleCtrl.GetRoleList)
		// 根据ID获取单个角色接口
		roleGroup.GET("/:id", middleware.PermissionMiddleware(), roleCtrl.GetRole)
		// 根据ID更新角色接口
		roleGroup.PUT("/:id", middleware.PermissionMiddleware(), roleCtrl.UpdateRole)
		// 根据ID删除角色接口
		roleGroup.DELETE("/:id", middleware.PermissionMiddleware(), roleCtrl.DeleteRole)
	}
}
