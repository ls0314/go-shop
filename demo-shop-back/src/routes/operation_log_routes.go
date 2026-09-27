package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"

	"github.com/gin-gonic/gin"
)

var operationLogCtrl *handler.OperationLogHandler

func InitOperationLogModule() {
	operationLogCtrl = handler.NewOperationLogHandler()
}

func RegisterOperationLogRoutes(r *gin.Engine) {
	operationLogGroup := r.Group("/api/v1/admin/platform/log")
	operationLogGroup.Use(middleware.AuthMiddleware())
	{
		operationLogGroup.GET("", middleware.PermissionMiddleware(), operationLogCtrl.GetOperationLogList)

	}
}
