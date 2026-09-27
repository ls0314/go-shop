package routes

import (
	"demo-shop-back/src/handler"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/service"

	"github.com/gin-gonic/gin"
)

// deptCtrl 部门模块全局控制器实例
// 作用：供路由注册时使用，持有部门Handler单例对象
var deptCtrl *handler.DeptHandler

// InitDeptModule 初始化部门模块
// 功能：完成部门模块 仓库层 → 服务层 → 控制层 的依赖注入与实例化
// 执行顺序：创建数据访问层实例 → 创建业务逻辑层实例 → 创建控制器实例
func InitDeptModule(deps service.ServiceDeps) {

	// 初始化部门控制器，赋值给全局控制器变量
	deptCtrl = handler.NewDeptHandler(deps)

}

// RegisterDeptRoutes 注册部门模块路由
// 参数：r *gin.Engine Gin路由引擎实例
// 功能：注册部门相关API路由，统一前缀 /api/v1/dept，并添加登录认证中间件
func RegisterDeptRoutes(r *gin.Engine) {
	// 创建部门接口路由分组，统一前缀 /api/v1/dept
	deptGroup := r.Group("/api/v1/dept")
	// 添加全局认证中间件（必须登录才能访问部门接口）
	deptGroup.Use(middleware.AuthMiddleware())
	{
		// 创建部门接口
		deptGroup.POST("", middleware.PermissionMiddleware(), deptCtrl.CreateDept)
		// 分页获取部门列表接口
		deptGroup.GET("", middleware.PermissionMiddleware(), deptCtrl.GetDeptList)
		// 根据ID获取单个部门接口
		deptGroup.GET("/:id", middleware.PermissionMiddleware(), deptCtrl.GetDept)
		// 根据用户Id构建该用户的部门树
		deptGroup.GET("/tree/:userId", middleware.PermissionMiddleware(), deptCtrl.GetDeptTreeByUserId)
		// 根据ID更新部门接口
		deptGroup.PUT("/:id", middleware.PermissionMiddleware(), deptCtrl.UpdateDept)
		// 根据ID删除部门接口
		deptGroup.DELETE("/:id", middleware.PermissionMiddleware(), deptCtrl.DeleteDept)
	}
}
