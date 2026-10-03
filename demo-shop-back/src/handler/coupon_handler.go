package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CouponHandler 优惠券模块 HTTP handler。
// 券域读写已迁 marketing-service(独立库 marketing_db),本层只做
// "HTTP 入参绑定 → RPC → HTTP 出参",不再直连券表、也不持有闸门缓存。
type CouponHandler struct {
	CouponService *service.CouponService // 优惠券服务层对象指针(RPC 门面)
}

// NewCouponHandler 新建优惠券模块的 HTTP handler 实例
// 接收值：deps - 服务层依赖(CouponRPC 未连上时为 nil)
// 返回值：*CouponHandler - 优惠券 handler 指针
func NewCouponHandler(deps service.ServiceDeps) *CouponHandler {
	return &CouponHandler{
		CouponService: service.NewCouponService(deps),
	}
}

// CreateCouponTemplate 创建优惠券模板接口
// 路由映射：POST /api/v1/admin/coupons
// 所需权限：platform:coupon:create（AuthMiddleware + PermissionMiddleware）
// 功能：接收前端传递的模板参数，绑定 JSON 后经 RPC 交给 marketing-service 创建
// 错误：参数绑定失败返回 400；业务失败（参数非法/有效期配置非法）返回 500；
// 下游不可用返回 503（failRPC 统一收敛）
func (ch *CouponHandler) CreateCouponTemplate(c *gin.Context) {
	var req requset.CreateCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := ch.CouponService.CreateCoupon(req)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}

// GetCouponList 优惠券模板列表接口
// 路由映射：GET /api/v1/admin/coupons
// 所需权限：platform:coupon:view（AuthMiddleware + PermissionMiddleware）
// 功能：绑定分页与筛选参数，经 RPC 分页查询模板
// 说明：分页归一化与实际生效值由服务端回显，本层不改写请求参数
func (ch *CouponHandler) GetCouponList(c *gin.Context) {
	var req requset.GetCouponListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	resp, err := ch.CouponService.GetCouponList(req)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}

// GetUserCouponList 用户优惠券列表接口
// 路由映射：GET /api/v1/coupons
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：从上下文获取当前用户ID（防越权），经 RPC 分页查询该用户的优惠券
// 错误：上下文无用户信息返回 500（UserInfoError）
func (ch *CouponHandler) GetUserCouponList(c *gin.Context) {
	var req requset.UserGetCouponListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}
	resp, err := ch.CouponService.UserGetCouponList(userId, req)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}

// ReceiveCoupon 用户领取优惠券接口
// 路由映射：POST /api/v1/coupons/receive/:id
// 鉴权：JWT（AuthMiddleware 注入用户信息）+ PerIPRateLimit
// 功能：解析路径参数 templateId，从上下文获取用户ID，经 RPC 领取
// 错误：路径参数非数字返回 400；业务失败（模板不存在/已领完/已达上限）返回 500
func (ch *CouponHandler) ReceiveCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}

	resp, err := ch.CouponService.ReceiveCoupon(userId, id)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}

// GetReceiveCouponList 领券中心模板列表接口
// 路由映射：GET /api/v1/coupons/templates
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：从上下文获取当前用户ID（算 held_count 用），经 RPC 分页查询可领取的模板
// 错误：上下文无用户信息返回 500（UserInfoError）
func (ch *CouponHandler) GetReceiveCouponList(c *gin.Context) {
	var req requset.UserGetTemplateListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}
	resp, err := ch.CouponService.GetReceiveCouponList(userId, req)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}

// GetAvailableCouponList 结算可用优惠券接口
// 路由映射：GET /api/v1/coupons/available?order_amount=xxx
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：绑定订单金额参数，从上下文获取用户ID，返回按实付升序的可用券列表
// 错误：参数绑定失败返回 400；order_amount 非法返回 500
func (ch *CouponHandler) GetAvailableCouponList(c *gin.Context) {
	var req requset.GetAvailableCouponReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}

	resp, err := ch.CouponService.GetAvailableCouponList(userId, req)
	if err != nil {
		failRPC(c, err)
		return
	}
	utils.Success(c, resp)
}
