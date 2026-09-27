package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var userDeptCtrl *handler.UserDeptHandler

func InitUserDeptModule(deps service.ServiceDeps) {
	userDeptCtrl = handler.NewUserDeptHandler(deps)
}

func RegisterUserDeptRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	userDeptGroup := r.Group("/api/v1/user")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	userDeptGroup.Use(middleware.AuthMiddleware())
	userDeptGroup.Use(middleware.OperationLogMiddleware(model.LogModuleUser))
	{
		// 批量创建用户部门关联接口
		userDeptGroup.POST("/assign-dept", permMW, userDeptCtrl.CreateUserDeptRel)
		// 获取某用户全部部门关联列表接口
		userDeptGroup.GET("/:id/dept", permMW, userDeptCtrl.GetUserDeptRelList)
		// 根据ID删除某用户全部部门关联接口
		userDeptGroup.DELETE("/:id/clear-dept", permMW, userDeptCtrl.DeleteUserAllDeptRel)
	}
}
