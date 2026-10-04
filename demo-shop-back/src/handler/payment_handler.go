package handler

import (
	"demo-shop-back/src/infra/tradeclient"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付管理handler层实例。
//
// 支付流水表已迁至 trade-service(独立库 trade_db),本层只做
// "HTTP 入参绑定 → RPC → HTTP 出参",不再直连 user_payment_record。
type PaymentHandler struct {
	tradeRPC *tradeclient.TradeClient
}

// NewPaymentHandler 创建支付管理handler层实例
// 接收值：无接收值，全局实例化
// 返回值：*PaymentHandler - 支付handler指针
func NewPaymentHandler(deps service.ServiceDeps) *PaymentHandler {
	return &PaymentHandler{
		tradeRPC: deps.TradeRPC,
	}
}

// CreatePayment 发起支付接口（用户端）
// 路由映射：POST /api/v1/user/pay/order/:orderId
// 功能：对指定订单发起支付。校验订单归属和状态后创建支付记录，返回支付流水号。
//
//	mock模式下返回支付记录信息，前端后续调回调接口完成支付；
//	wechat/alipay模式下返回支付链接/二维码。
//
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、请求体、返回响应
// 路径参数：
//
//	orderId - 订单ID，int64类型
//
// 请求参数（Body）：
//
//	pay_method - 支付方式，string类型，必填（mock/wechat/alipay）
//
// 响应：
//
//	400：参数绑定失败/订单ID格式错误/用户未登录
//	500：服务层处理失败（10001订单不存在或不属于当前用户/10002状态不允许/10003已有进行中的支付）
//	200：发起成功，返回支付记录（含流水号pay_no、金额、状态；真实支付额外返回pay_url/qr_code）
func (p *PaymentHandler) CreatePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
	}

	var req requset.CreatePaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, errMsg, err := p.tradeRPC.CreatePayment(id, userId, req.PayMethod)
	if err != nil {
		failTradeRPC(c, err)
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	utils.Success(c, resp)
}

// GetPayment 查询支付状态接口（用户端）
// 路由映射：GET /api/v1/user/pay/:payNo
// 功能：根据支付流水号查询支付记录和当前状态。校验归属后，联表查询订单号一并返回
// 参数：c *gin.Context Gin上下文，用于获取URL参数、当前用户、返回响应
// 路径参数：
//
//	payNo - 支付流水号，string类型（如 PAY20260728143052B7D2）
//
// 响应：
//
//	400：用户未登录
//	500：服务层查询失败（10004支付记录不存在/10005无权查看）
//	200：查询成功，返回支付详情（含订单号order_no、支付方式、金额、状态、支付时间、第三方交易号）
func (p *PaymentHandler) GetPayment(c *gin.Context) {
	payNo := c.Param("payNo")

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, errMsg, err := p.tradeRPC.GetPayment(userId, payNo)
	if err != nil {
		failTradeRPC(c, err)
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	utils.Success(c, resp)
}

// MockCallback 模拟支付回调接口
// 路由映射：POST /api/v1/pay/callback/mock
// 鉴权：无（无需登录，便于开发测试；正式环境由支付平台回调，亦无JWT）
// 功能：模拟支付网关回调，完成支付。mock模式下前端发起支付后主动调用此接口，
//
//	后端执行完整的安全校验链（幂等→金额→状态）后在事务内完成：
//	支付记录→success、订单→paid、扣减锁定库存、累加销量。
//
// 参数：c *gin.Context Gin上下文，用于获取请求体、返回响应
// 请求参数（Body）：
//
//	pay_no   - 支付流水号，string类型，必填
//	trade_no  - 模拟第三方交易号，string类型，可选
//
// 响应：
//
//	400：参数绑定失败
//	500：回调处理失败（10004支付记录不存在/10006状态不正确/10007金额不匹配）
//	200：支付成功
func (p *PaymentHandler) MockCallback(c *gin.Context) {
	var req requset.CallbackPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	// **错误语义与其它接口不同**:这里 err != nil 也可能表示
	// "支付已成功但后续步骤(订单推进/扣库存)失败,请重试回调" ——
	// 那是向前补偿的信号,不是回调无效。见 payment.proto 的说明。
	// 故此处用 failTradeRPC(区分依赖不可用)而不是一律 500。
	_, errMsg, err := p.tradeRPC.HandleCallback(req.PayNo, req.TradeNo)
	if err != nil {
		failTradeRPC(c, err)
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	utils.Success(c, nil)
}

// GetPaymentList 管理端支付列表接口
// 路由映射：GET /api/v1/admin/pay/list
// 鉴权：platform:pay:view
// 功能：分页查看所有支付记录，支持按支付状态/支付方式/订单号/时间范围筛选。
//
//	联表查询用户名和订单号，按创建时间倒序。
//
// 参数：c *gin.Context Gin上下文，用于获取查询参数、返回响应
// 查询参数：
//
//	page       - 页码，int类型，默认1
//	page_size  - 每页条数，int类型，默认20，最大100
//	pay_status  - 支付状态筛选，string类型，可选（pending/success/failed/closed）
//	pay_method  - 支付方式筛选，string类型，可选（mock/wechat/alipay）
//	order_no    - 订单号搜索，string类型，可选
//	start_time  - 开始时间，string类型，可选（yyyy-MM-dd）
//	end_time    - 结束时间，string类型，可选（yyyy-MM-dd）
//
// 响应：
//
//	400：参数绑定失败
//	500：服务层查询失败
//	200：查询成功，返回分页列表（含用户名username、订单号order_no）
func (p *PaymentHandler) GetPaymentList(c *gin.Context) {
	var req requset.GetPaymentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, errMsg, err := p.tradeRPC.GetPaymentList(req.Page, req.PageSize, req.PayStatus,
		req.PayMethod, req.OrderNo, req.StartTime, req.EndTime)
	if err != nil {
		failTradeRPC(c, err)
		return
	}
	if errMsg != "" {
		utils.Error(c, 500, errMsg)
		return
	}
	utils.Success(c, resp)
}
