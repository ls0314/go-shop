package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductHandler 商品模块 HTTP handler
type ProductHandler struct {
	ProductService *service.ProductService // 商品服务层对象指针
}

// NewProductHandler 新建商品模块的 HTTP handler 实例
// 接收值：无接收值，全局实例化
// 返回值：*ProductHandler - 商品 handler 指针
func NewProductHandler(deps service.ServiceDeps) *ProductHandler {
	return &ProductHandler{
		ProductService: service.NewProductService(deps),
	}
}

// CreateProduct 创建商品（SPU）接口
// 路由映射：POST /api/v1/platform/product
// 所需权限：platform:product:create
// 功能：接收前端传递的商品 SPU 信息，校验参数后调用服务层创建商品
// 参数：c *gin.Context Gin上下文，用于接收请求参数、返回响应
// 请求参数：
//
//	spu_name      - 商品名称
//	category_id   - 类目ID
//	description   - 商品描述
//	spu_image     - 商品图片
//	status        - 状态
//	publish_status - 上架状态
//	sort_order    - 排序
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层创建商品失败，返回服务器异常信息
//	200：创建成功，返回新创建商品的 spu_id
func (p *ProductHandler) CreateProduct(c *gin.Context) {
	var product model.SysProductSpu

	if err := c.ShouldBindJSON(&product); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	spuId, err := p.ProductService.CreateProduct(&product)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"spu_id": spuId})
}

// GetProductList 分页查询商品列表接口（管理端）
// 路由映射：GET /api/v1/platform/product
// 所需权限：platform:product:list
// 功能：支持分页和条件筛选查询商品 SPU 列表，返回分页数据和总条数
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 请求参数：
//
//	page        - 页码，默认值1
//	pageSize    - 每页条数，默认值10
//	spu_name    - 商品名称（模糊搜索）
//	category_id - 类目ID
//	status      - 状态
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层查询商品列表失败
//	200：查询成功，返回商品列表和分页信息
func (p *ProductHandler) GetProductList(c *gin.Context) {
	var req requset.SpuQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	productList, err := p.ProductService.GetProductSpuList(req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, productList)
}

// GetProduct 查询商品信息接口（管理端，根据ID查询）
// 路由映射：GET /api/v1/platform/product/:id
// 所需权限：platform:product:view
// 功能：从URL路径中获取商品 SPU ID，查询并返回对应商品的完整详情（含 SKU 列表）
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询商品失败
//	200：查询成功，返回商品完整信息（含 SKU 列表）
func (p *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	product, err := p.ProductService.GetProduct(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, product)
}

// UpdateProduct 局部更新商品接口
// 路由映射：PUT /api/v1/platform/product/:id
// 所需权限：platform:product:update
// 功能：从URL获取商品ID，接收前端传入的更新字段，执行商品信息局部更新，返回更新后的商品详情
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//	请求体为需要更新的字段（map格式，支持部分字段更新）
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：更新商品信息失败
//	200：更新成功，返回更新后的商品完整信息
func (p *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	var product map[string]interface{}
	if err := c.ShouldBindJSON(&product); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	updateProduct, err := p.ProductService.UpdateProduct(id, product)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, updateProduct)
}

// UpdateFullProduct 全量更新商品接口
// 路由映射：PUT /api/v1/platform/product/:id/full
// 所需权限：platform:product:update
// 功能：从URL获取商品ID，接收前端传入的完整商品信息（含 SKU 列表），执行商品全量更新
// 参数：c *gin.Context Gin上下文，用于获取URL参数、接收请求体、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//	请求体为完整的商品信息，包含 SPU 字段和 SKU 列表
//
// 响应：
//
//	400：ID格式错误 / 请求参数绑定失败
//	500：全量更新商品信息失败
//	200：更新成功，返回更新后的商品完整信息
func (p *ProductHandler) UpdateFullProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	var product requset.FullUpdateProductReq
	if err := c.ShouldBindJSON(&product); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	updateProduct, err := p.ProductService.UpdateProductFull(id, product)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, updateProduct)
}

// DeleteProduct 删除商品接口（根据ID删除）
// 路由映射：DELETE /api/v1/platform/product/:id
// 所需权限：platform:product:delete
// 功能：从URL路径获取商品ID，调用服务层执行删除操作，同时删除关联的 SKU 数据
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层删除商品失败
//	200：删除成功，返回空数据
func (p *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := p.ProductService.DeleteProduct(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// PublishProduct 上架商品接口
// 路由映射：PUT /api/v1/platform/product/:id/publish
// 所需权限：platform:product:publish
// 功能：从URL路径获取商品ID，调用服务层将商品发布状态修改为上架
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层上架商品失败
//	200：上架成功，返回空数据
func (p *ProductHandler) PublishProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := p.ProductService.PublishProduct(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// WithdrawnProduct 下架商品接口
// 路由映射：PUT /api/v1/platform/product/:id/withdraw
// 所需权限：platform:product:withdraw
// 功能：从URL路径获取商品ID，调用服务层将商品发布状态修改为下架
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//
// 响应：
//
//	400：ID格式错误/不存在
//	500：服务层下架商品失败
//	200：下架成功，返回空数据
func (p *ProductHandler) WithdrawnProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	if err := p.ProductService.WithdrawProduct(id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// UserProductList 分页查询商品列表接口（用户端）
// 路由映射：GET /api/v1/product
// 功能：支持分页和条件筛选查询已上架的商品 SPU 列表，仅返回用户端可见字段
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 请求参数：
//
//	page        - 页码，默认值1
//	pageSize    - 每页条数，默认值10
//	spu_name    - 商品名称（模糊搜索）
//	category_id - 类目ID
//
// 响应：
//
//	400：请求参数绑定失败，返回参数错误信息
//	500：服务层查询商品列表失败
//	200：查询成功，返回商品列表和分页信息
func (p *ProductHandler) UserProductList(c *gin.Context) {
	var req requset.SpuQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	productList, err := p.ProductService.UserGetProductSpuList(req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, productList)
}

// UserProduct 查询商品详情接口（用户端，根据ID查询）
// 路由映射：GET /api/v1/product/:id
// 功能：从URL路径中获取商品 SPU ID，查询并返回对应用户端可见的商品详情（含 SKU 列表）
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 请求参数：
//
//	id - URL路径参数，商品 SPU ID
//
// 响应：
//
//	400：URL参数ID格式错误/不存在
//	500：服务层查询商品失败
//	200：查询成功，返回用户端商品详情（含 SKU 列表）
func (p *ProductHandler) UserProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	product, err := p.ProductService.UserGetProduct(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, product)
}
