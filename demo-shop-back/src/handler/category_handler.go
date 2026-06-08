package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 类目表handler层实例
type CategoryHandler struct {
	CategoryService *service.CategoryService // 类目服务层对象指针
}

// NewCategoryHandler 新建类目表中的HTTP handler实例
// 接收值：无接收值，全局实例化
// 返回值：*CategoryHandler - 类目handler指针
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{
		CategoryService: service.NewCategoryService(),
	}
}

// CreateCategory 创建类目接口
// 路由映射：POST /api/v1/platform/category
// 所需权限：platform:category:create
// 功能：接收前端传递的类目信息，校验参数后调用服务层创建类目，自动计算类目level和path
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	category_name - 类目名称
//	parent_id     - 父类目ID，一级类目传0
//	sort_order    - 排序
//	icon_url      - 图标URL
//	is_visible    - 是否可见
//	status        - 状态
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建类目失败，返回服务器异常信息
//	200：创建成功，返回创建完成的类目完整信息
func (cg *CategoryHandler) CreateCategory(c *gin.Context) {
	var category model.SysCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	categoryData, err := cg.CategoryService.CreateCategory(&category)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, categoryData)
}

// GetCategory 查询类目信息接口（根据ID查询）
// 路由映射：GET /api/v1/platform/category/:id
// 所需权限：platform:category:view
// 功能：从URL路径中获取类目ID，查询并返回对应类目详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，类目ID
//
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询类目失败
//	200：查询成功，返回类目详细信息
func (cg *CategoryHandler) GetCategory(c *gin.Context) {
	// 通过传入URL地址获取INT格式的ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	category, err := cg.CategoryService.GetCategory(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, category)
}

// GetCategoryTree 获取类目树接口
// 路由映射：GET /api/v1/platform/category/tree
// 所需权限：platform:category:tree
// 功能：根据查询参数获取类目树，可按层级查询，并可控制是否包含禁用类目
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 请求参数：
//
//	level            - 查询层级，默认值0，表示查询全部或从根层级开始
//	include_disabled - 是否包含禁用类目，默认值false
//
// 响应：
//
//	500：服务层查询类目树失败
//	200：查询成功，返回类目树数据
func (cg *CategoryHandler) GetCategoryTree(c *gin.Context) {
	level, _ := strconv.ParseInt(c.DefaultQuery("level", "0"), 10, 64)
	includeDisabled, _ := strconv.ParseBool(c.DefaultQuery("include_disabled", "false"))

	categoryTree, err := cg.CategoryService.GetCategoryTree(level, includeDisabled)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, categoryTree)
}

// GetCategoryList 分页查询类目信息接口
// 路由映射：GET /api/v1/platform/category
// 功能：支持分页查询类目列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取分页参数、返回响应
// 请求参数：
//
//	page     - 页码，默认值1
//	pageSize - 每页条数，默认值10
//
// 响应：
//
//	500：服务层查询类目列表失败
//	200：查询成功，返回类目列表、总条数、当前页码、每页条数
func (cg *CategoryHandler) GetCategoryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	categoryList, total, err := cg.CategoryService.GetCategoryList(page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"list":     categoryList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetCategoryChildrenList 获取子类目列表接口
// 路由映射：GET /api/v1/platform/category/children/:id
// 所需权限：platform:category:children
// 功能：从URL路径中获取父类目ID，查询并返回该类目下的直接子类目列表
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，父类目ID
//
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询子类目列表失败
//	200：查询成功，返回子类目列表
func (cg *CategoryHandler) GetCategoryChildrenList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	childrenList, err := cg.CategoryService.GetCategoryChildrenList(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, childrenList)
}

// UpdateCategory 更新类目接口
// 路由映射：PUT /api/v1/platform/category/:id
// 所需权限：platform:category:update
// 功能：从URL获取类目ID，接收前端传入的更新字段，执行类目信息局部更新，返回更新后的类目详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id             - URL路径参数，类目ID
//	category_name  - 类目名称
//	sort_order     - 排序
//	icon_url       - 图标URL
//	is_visible     - 是否可见
//	status         - 状态
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新类目信息失败 / 查询更新后类目信息失败
//	200：更新成功，返回更新后的完整类目信息
func (cg *CategoryHandler) UpdateCategory(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	var category map[string]interface{}
	if err := c.ShouldBind(&category); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	categoryData, err := cg.CategoryService.UpdateCategory(id, category)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, categoryData)

}

// DeleteCategory 删除类目接口（根据ID删除）
// 路由映射：DELETE /api/v1/platform/category/:id
// 所需权限：platform:category:delete
// 功能：从URL路径获取类目ID，调用服务层执行删除操作，删除前需校验类目是否存在、是否存在子类目或关联商品
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，类目ID
//
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除类目失败
//	200：删除成功，返回空数据
func (cg *CategoryHandler) DeleteCategory(c *gin.Context) {
	// 通过URL地址获取所删除ID的INT格式
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	if err := cg.CategoryService.DeleteCategory(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
