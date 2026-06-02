package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var userCtrl *handler.UserHandler

func InitUserModule() {
	userCtrl = handler.NewUserHandler()
}

func RegisterUserRoutes(r *gin.Engine) {
	userPublic := r.Group("/api/v1/user")
	{
		userPublic.POST("/register", userCtrl.CreateUserHandler)
		userPublic.POST("/login", userCtrl.LoginHandler)
		userPublic.POST("/refresh", userCtrl.RefreshHandler)
	}

	userPrivate := r.Group("/api/v1/user")
	userPrivate.Use(middleware.AuthMiddleware())
	{
		userPrivate.GET("/info", userCtrl.GetUserInfo)
		userPrivate.GET("", userCtrl.GetUserList)
		userPrivate.POST("", userCtrl.CreateUserHandler)
		userPrivate.PUT("/:id", userCtrl.UpdateUser)
		userPrivate.DELETE("/:id", userCtrl.DeleteUser)
		userPrivate.GET("/:id", userCtrl.GetUser)
	}

}
