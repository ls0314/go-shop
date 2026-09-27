package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var userCtrl *handler.UserHandler

func InitUserModule(deps service.ServiceDeps) {
	userCtrl = handler.NewUserHandler(deps)
}

func RegisterUserRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	userPublic := r.Group("/api/v1/user")
	{
		userPublic.POST("/register", middleware.PerIPRateLimit(), userCtrl.CreateUserHandler)
		userPublic.POST("/login", middleware.PerIPRateLimit(), userCtrl.LoginHandler)
		userPublic.POST("/refresh", userCtrl.RefreshHandler)
	}

	userPrivate := r.Group("/api/v1/user")
	userPrivate.Use(middleware.AuthMiddleware())
	userPrivate.Use(middleware.OperationLogMiddleware(model.LogModuleUser))
	{
		userPrivate.GET("/info", userCtrl.GetUserInfo)
		userPrivate.GET("/perms", userCtrl.GetUserPerms)
		userPrivate.GET("", permMW, userCtrl.GetUserList)
		userPrivate.POST("", permMW, userCtrl.CreateUserHandler)
		userPrivate.PUT("/:id", permMW, userCtrl.UpdateUser)
		userPrivate.DELETE("/:id", permMW, userCtrl.DeleteUser)
		userPrivate.GET("/:id", permMW, userCtrl.GetUser)
	}

}
