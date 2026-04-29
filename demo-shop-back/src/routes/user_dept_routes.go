package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var userDeptCtrl *handler.UserDeptHandler

func InitUserDeptModule() {
	userDeptCtrl = handler.NewUserDeptHandler()
}

func RegisterUserDeptRoutes(r *gin.Engine) {
	// 创建角色接口路由分组，统一前缀 /api/v1/role
	userDeptGroup := r.Group("/api/v1/user")
	// 添加全局认证中间件（必须登录才能访问角色接口）
	userDeptGroup.Use(middleware.AuthMiddleware())
	{
		// 批量创建用户部门关联接口
		userDeptGroup.POST("/assign-dept", userDeptCtrl.CreateUserDeptRel)
		// 获取某用户全部部门关联列表接口
		userDeptGroup.GET("/:id/dept", userDeptCtrl.GetUserDeptRelList)
		// 根据ID删除某用户全部部门关联接口
		userDeptGroup.DELETE("/:id/clear-dept", userDeptCtrl.DeleteUserAllDeptRel)
	}
}
