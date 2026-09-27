package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单管理handler层实例
type OrderHandler struct {
	OrderService *service.OrderService // 订单服务层对象指针
}

// NewOrderHandler 创建订单管理handler层实例
// 接收值：无接收值，全局实例化
// 返回值：*OrderHandler - 订单handler指针
func NewOrderHandler(deps service.ServiceDeps) *OrderHandler {
	return &OrderHandler{
		OrderService: service.NewOrderService(deps),
	}
}

// CreateOrder 创建订单接口（用户端）
// 路由映射：POST /api/v1/user/platform/orders
// 功能：从购物车选中项创建订单。校验地址归属、商品状态和库存后，在事务内写入订单主表+明细+日志，锁定库存并清理购物车
// 参数：c *gin.Context Gin上下文，用于获取当前用户、接收请求体、返回响应
// 请求参数：
//
//	address_id     - 收货地址ID，int64类型，必填
//	idempotent_key - 幂等键（UUID），string类型，必填，重复提交返回已有订单
//	buyer_remark   - 买家备注，string类型，最长200字符，可选
//
// 响应：
//
//	400：参数绑定失败/用户未登录
//	500：服务层创建失败（9001地址不存在/9002无结算商品/9003已下架或库存不足）
//	200：创建成功，返回订单号、金额、状态、支付截止时间
func (o *OrderHandler) CreateOrder(c *gin.Context) {
	var orderReq requset.CreatOrderReq
	if err := c.ShouldBindJSON(&orderReq); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, userName, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.CreateOrder(&orderReq, userId, userName)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserOrderList 用户订单列表接口
// 路由映射：GET /api/v1/user/platform/orders
// 功能：分页查询当前用户订单，支持按状态筛选，按下单时间倒序
// 参数：c *gin.Context Gin上下文，用于获取当前用户、查询参数、返回响应
// 查询参数：
//
//	page         - 页码，int类型，默认1
//	page_size    - 每页条数，int类型，默认10
//	order_status - 状态筛选，string类型，可选（pending_pay/paid/shipped/completed/cancelled）
//
// 响应：
//
//	400：参数绑定失败/用户未登录
//	500：服务层查询失败
//	200：查询成功，返回订单列表+分页信息（含detail_count/first_image冗余字段）
func (o *OrderHandler) GetUserOrderList(c *gin.Context) {
	var req requset.UserGetOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.GetUserOrderList(userId, req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserOrderInfo 用户订单详情接口
// 路由映射：GET /api/v1/user/platform/orders/:id
// 功能：获取订单完整信息，含明细列表、地址快照、订单日志。校验订单归属
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 订单ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层查询失败（9005订单不存在/9006无权查看）
//	200：查询成功，返回订单详情+明细列表+日志列表
func (o *OrderHandler) GetUserOrderInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.GetUserOrder(id, userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// CancelOrder 取消订单接口（用户端）
// 路由映射：PUT /api/v1/user/platform/orders/:id/cancel
// 功能：取消待支付订单。校验归属和状态后，在事务内更新状态+写日志+释放锁定库存
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 订单ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层取消失败（9005不存在/9006无权操作/9007状态不允许）
//	200：取消成功，返回订单ID+单号+状态
func (o *OrderHandler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, userName, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.CancelOrder(id, userId, userName)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// GetOrderList 管理端订单列表接口
// 路由映射：GET /api/v1/platform/orders
// 功能：分页查询所有订单，支持按状态/订单号/时间范围筛选，联表查询用户信息和收货人
// 鉴权：platform:order:view
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 查询参数：
//
//	page         - 页码，int类型，默认1
//	page_size    - 每页条数，int类型，默认10
//	order_status - 状态筛选，string类型，可选
//	order_no     - 订单号精确搜索，string类型，可选
//	start_time   - 下单开始日期，string类型，可选（yyyy-MM-dd）
//	end_time     - 下单结束日期，string类型，可选（yyyy-MM-dd）
//
// 响应：
//
//	400：参数绑定失败
//	500：服务层查询失败
//	200：查询成功，返回订单列表（含用户名、收货人姓名电话）
func (o *OrderHandler) GetOrderList(c *gin.Context) {
	var req requset.GetOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := o.OrderService.GetOrderList(req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetOrderInfo 管理端订单详情接口
// 路由映射：GET /api/v1/platform/orders/:id
// 功能：获取任意订单完整信息，含明细列表、地址快照、订单日志，额外返回下单用户名和用户ID
// 鉴权：platform:order:view（无归属校验，可查看任意用户订单）
// 参数：c *gin.Context Gin上下文，用于获取URL参数、返回响应
// 路径参数：
//
//	id - 订单ID，int64类型
//
// 响应：
//
//	400：ID格式错误
//	500：服务层查询失败（9005订单不存在）
//	200：查询成功，在用户端详情字段基础上额外返回username+user_id
func (o *OrderHandler) GetOrderInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := o.OrderService.GetOrder(id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// OrderShip 订单发货接口（管理端）
// 路由映射：PUT /api/v1/platform/orders/:id/ship
// 功能：对已支付订单执行发货。校验快递信息完整性和订单状态后，在事务内更新状态+写日志
// 鉴权：platform:order:ship
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前管理员、请求体、返回响应
// 路径参数：
//
//	id - 订单ID，int64类型
//
// 请求参数（Body）：
//
//	express_company - 快递公司名称，string类型，必填
//	tracking_no    - 快递单号，string类型，必填
//
// 响应：
//
//	400：ID格式错误/参数绑定失败/用户未登录
//	500：服务层发货失败（9005不存在/9008状态不允许/9009快递信息不完整）
//	200：发货成功，返回订单状态+快递信息
func (o *OrderHandler) OrderShip(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	var req requset.OrderShipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	_, userName, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.OrderShip(id, userName, req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// ConfirmOrder 确认收货接口（用户端）
// 路由映射：PUT /api/v1/user/platform/orders/:id/confirm
// 功能：用户确认收到商品。校验归属和状态后，在事务内更新状态为completed+写日志
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	id - 订单ID，int64类型
//
// 响应：
//
//	400：ID格式错误/用户未登录
//	500：服务层确认失败（9005不存在/9006无权操作/9010状态不允许）
//	200：确认成功，返回订单ID+单号+状态
func (o *OrderHandler) ConfirmOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, userName, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resp, err := o.OrderService.ConfirmOrder(id, userId, userName)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
