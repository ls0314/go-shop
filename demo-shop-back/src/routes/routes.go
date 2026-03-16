package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("api/v1")
	{
		user := v1.Group("user")
		{
			user.POST("/register", handler.RegisterHandler)
			user.POST("/login", handler.LoginHandler)
			user.Use(middleware.AuthMiddleware(middleware.InitJWT("demo_shop")))
			user.GET("/info", handler.GetUserInfo)
		}
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	return r
}
