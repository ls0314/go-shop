package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var operationLogCtrl *handler.OperationLogHandler

func InitOperationLogModule(deps service.ServiceDeps) {
	operationLogCtrl = handler.NewOperationLogHandler(deps)
}

func RegisterOperationLogRoutes(r *gin.Engine) {
	operationLogGroup := r.Group("/api/v1/admin/platform/log")
	operationLogGroup.Use(middleware.AuthMiddleware())
	{
		operationLogGroup.GET("", middleware.PermissionMiddleware(), operationLogCtrl.GetOperationLogList)

	}
}
