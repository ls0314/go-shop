package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/repository"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

var operationLogCtrl *handler.OperationLogHandler

func InitOperationLogModule(deps service.ServiceDeps) {
	operationLogCtrl = handler.NewOperationLogHandler(deps)
}

func RegisterOperationLogRoutes(r *gin.Engine, deps service.ServiceDeps) {
	// 权限中间件:装配期构造一次,组内所有路由复用同一个闭包
	permMW := middleware.PermissionMiddleware(repository.NewPermissionRepo(deps.DB), deps.Cache)
	operationLogGroup := r.Group("/api/v1/admin/platform/log")
	operationLogGroup.Use(middleware.AuthMiddleware())
	{
		operationLogGroup.GET("", permMW, operationLogCtrl.GetOperationLogList)

	}
}
