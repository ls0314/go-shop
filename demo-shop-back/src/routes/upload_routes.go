package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

// uploadCtrl 上传模块全局控制器实例
// 作用：供路由注册时使用，持有上传Handler单例对象
var uploadCtrl *handler.UploadHandler

// InitUploadModule 初始化上传模块
// 功能：完成上传模块 服务层 → 控制层 的依赖注入与实例化
func InitUploadModule(deps service.ServiceDeps) {
	uploadCtrl = handler.NewUploadHandler(deps)
}

// RegisterUploadRoutes 注册上传模块路由
// 参数：r *gin.Engine Gin路由引擎实例
// 功能：注册切片上传相关API路由，统一前缀 /api/v1/upload，并添加登录认证中间件
// 注册路由：
//
//	POST /api/v1/upload/chunk - 上传文件分片（所有分片上传完毕后自动合并）
func RegisterUploadRoutes(r *gin.Engine) {
	uploadGroup := r.Group("/api/v1/upload")
	uploadGroup.Use(middleware.AuthMiddleware())
	{
		uploadGroup.POST("/chunk", uploadCtrl.UploadChunk)
	}
}
