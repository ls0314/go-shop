package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CartItemHandler 购物车管理handler层实例
type CartItemHandler struct {
	CartItemService *service.CartItemService // 购物车服务层对象指针
}

// NewCartItemHandler 创建购物车管理handler层实例
// 接收值：无接收值，全局实例化
// 返回值：*CartItemHandler - 购物车handler指针
func NewCartItemHandler() *CartItemHandler {
	return &CartItemHandler{
		CartItemService: service.NewCartItemService(),
	}
}

// AddCartItem 加入购物车接口
// 路由映射：POST /api/v1/user/cart
// 功能：接收前端传入的SKU ID和数量，校验商品可售状态和库存后加入购物车。已存在则累加数量，首次加入则新增记录
// 参数：c *gin.Context Gin上下文，用于获取当前用户、接收请求体、返回响应
// 请求参数：
//
//	sku_id   - SKU ID，int64类型，必填
//	quantity - 购买数量，int64类型，可选（默认1，范围1-999）
//
// 响应：
//
//	400：请求参数绑定失败/用户未登录
//	500：服务层处理失败（8001 SKU不存在/8002 库存不足/8003 购物车上限/8004 数量超限）
//	200：加入成功，返回购物车项ID和总数量
func (ci *CartItemHandler) AddCartItem(c *gin.Context) {
	// 实例化后绑定请求参数
	var cartItem model.UserCartItem
	if err := c.ShouldBindJSON(&cartItem); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 若客户端未传user_id则注入JWT用户ID（防越权）
	if cartItem.UserId == 0 {
		cartItem.UserId = userId
	}
	// 调用服务层加入购物车（含商品校验、重复累加、上限检查）
	resp, err := ci.CartItemService.CreateCartItem(userId, &cartItem)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 加入成功，返回购物车项信息
	utils.Success(c, resp)
}

// GetCartItemList 获取购物车列表接口
// 路由映射：GET /api/v1/user/cart
// 功能：获取当前用户购物车全部记录，按更新时间倒序。关联查询SKU/SPU实时价格和库存，标记不可购买项
// 参数：c *gin.Context Gin上下文，用于获取当前用户、返回响应
// 响应：
//
//	400：用户未登录
//	500：服务层查询失败
//	200：查询成功，返回购物车数组（含可用状态、不可购买原因）
func (ci *CartItemHandler) GetCartItemList(c *gin.Context) {
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层获取购物车列表（含实时联表查询和可用性计算）
	resp, err := ci.CartItemService.GetCartItemList(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回购物车列表
	utils.Success(c, resp)
}

// UpdateCartItem 修改购物车项接口
// 路由映射：PUT /api/v1/user/cart/:id
// 功能：从URL获取购物车项ID，接收前端传入的新数量或选中状态，校验归属和库存后执行更新
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、接收请求体、返回响应
// 路径参数：
//
//	id - 购物车项ID，int64类型
//
// 请求参数（Body，可选但至少传一项）：
//
//	quantity    - 新数量，int64类型，范围1-999
//	is_selected - 是否选中，bool类型
//
// 响应：
//
//	400：ID格式错误/参数绑定失败/用户未登录
//	500：服务层更新失败（8002库存不足/8005不存在/8006无权操作）
//	200：更新成功，返回更新后的完整购物车项信息
func (ci *CartItemHandler) UpdateCartItem(c *gin.Context) {
	// 通过传入URL地址获取INT格式的购物车项ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 实例化后绑定更新参数
	var updateCartItem requset.UpdateCartItemReq
	if err := c.ShouldBindJSON(&updateCartItem); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层更新购物车项（含归属校验和库存校验）
	resp, err := ci.CartItemService.UpdateCartItem(id, userId, updateCartItem)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 更新成功，返回更新后完整信息
	utils.Success(c, resp)
}

// DeleteCartItem 删除购物车项接口
// 路由映射：DELETE /api/v1/user/cart/:id
// 功能：从URL获取购物车项ID，校验归属后从购物车移除
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 购物车项ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层删除失败（8005不存在/8006无权操作）
//	200：删除成功
func (ci *CartItemHandler) DeleteCartItem(c *gin.Context) {
	// 通过传入URL地址获取INT格式的购物车项ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层删除购物车项（含归属校验）
	if err := ci.CartItemService.DeleteCartItem(id, userId); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 删除成功
	utils.Success(c, nil)
}

// SelectAllCartItem 全选/取消全选接口
// 路由映射：PUT /api/v1/user/cart/select-all
// 功能：一键选中或取消选中当前用户所有购物车项
// 参数：c *gin.Context Gin上下文，用于获取当前用户、接收请求体、返回响应
// 请求参数：
//
//	is_selected - 是否全选，bool类型，必填
//
// 响应：
//
//	400：请求参数绑定失败/用户未登录
//	500：服务层操作失败
//	200：操作成功，返回受影响的记录数
func (ci *CartItemHandler) SelectAllCartItem(c *gin.Context) {
	// 实例化后绑定请求参数
	var req requset.SelectCartItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层执行全选/取消全选
	total, err := ci.CartItemService.SelectAllCartItem(userId, req.IsSelected)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 操作成功，返回受影响的记录数
	utils.Success(c, gin.H{
		"affected": total,
	})
}

// GetCartItemTotal 获取购物车总数量接口
// 路由映射：GET /api/v1/user/cart/count
// 功能：获取当前用户购物车中的总商品种类数（用于导航栏角标展示）
// 参数：c *gin.Context Gin上下文，用于获取当前用户、返回响应
// 响应：
//
//	400：用户未登录
//	500：服务层查询失败
//	200：查询成功，返回商品种类总数
func (ci *CartItemHandler) GetCartItemTotal(c *gin.Context) {
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层统计购物车项数量
	total, err := ci.CartItemService.GetCartItemTotal(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回总数
	utils.Success(c, gin.H{
		"count": total,
	})
}

// GetPayPreviewCartItem 选中项结算预览接口
// 路由映射：GET /api/v1/user/cart/preview
// 功能：获取当前用户所有选中项的汇总信息，含总金额、总数量、不可购买项列表
// 参数：c *gin.Context Gin上下文，用于获取当前用户、返回响应
// 响应：
//
//	400：用户未登录
//	500：服务层查询失败
//	200：查询成功，返回选中项列表及汇总数据（total_count/total_quantity/total_amount/has_unavailable）
func (ci *CartItemHandler) GetPayPreviewCartItem(c *gin.Context) {
	// 从JWT上下文获取当前用户ID
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 调用服务层获取选中项结算预览
	resp, err := ci.CartItemService.GetPayPreviewCartItem(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	// 查询成功，返回结算预览数据
	utils.Success(c, resp)
}
