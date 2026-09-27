package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var userCtrl *handler.UserHandler

func InitUserModule(deps service.ServiceDeps) {
	userCtrl = handler.NewUserHandler(deps)
}

func RegisterUserRoutes(r *gin.Engine) {
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
		userPrivate.GET("", middleware.PermissionMiddleware(), userCtrl.GetUserList)
		userPrivate.POST("", middleware.PermissionMiddleware(), userCtrl.CreateUserHandler)
		userPrivate.PUT("/:id", middleware.PermissionMiddleware(), userCtrl.UpdateUser)
		userPrivate.DELETE("/:id", middleware.PermissionMiddleware(), userCtrl.DeleteUser)
		userPrivate.GET("/:id", middleware.PermissionMiddleware(), userCtrl.GetUser)
	}

}
