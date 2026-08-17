package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CouponHandler 优惠券模块 HTTP handler
type CouponHandler struct {
	CouponService *service.CouponService // 优惠券服务层对象指针
}

// NewCouponHandler 新建优惠券模块的 HTTP handler 实例
// 接收值：无接收值，全局实例化
// 返回值：*CouponHandler - 优惠券 handler 指针
func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		CouponService: service.NewCouponService(),
	}
}

// CreateCouponTemplate 创建优惠券模板接口
// 路由映射：POST /api/v1/admin/platform/coupons
// 所需权限：platform:coupon:create（AuthMiddleware + PermissionMiddleware）
// 功能：接收前端传递的模板参数，绑定 JSON 后调用服务层创建
// 错误：参数绑定失败返回 400；服务层校验/落库失败返回 500（错误信息为业务错误码文案）
func (ch *CouponHandler) CreateCouponTemplate(c *gin.Context) {
	var req requset.CreateCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := ch.CouponService.CreateCoupon(req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetCouponList 优惠券模板列表接口
// 路由映射：GET /api/v1/admin/platform/coupons
// 所需权限：platform:coupon:view（AuthMiddleware + PermissionMiddleware）
// 功能：绑定分页与筛选参数，调用服务层分页查询模板
func (ch *CouponHandler) GetCouponList(c *gin.Context) {
	var req requset.GetCouponListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	resp, err := ch.CouponService.GetCouponList(req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserCouponList 用户优惠券列表接口
// 路由映射：GET /api/v1/users/platform/coupons
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：从上下文获取当前用户ID（防越权），分页查询该用户的优惠券
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
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// ReceiveCoupon 用户领取优惠券接口
// 路由映射：POST /api/v1/users/platform/coupons/receive/:templateId
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：解析路径参数 templateId，从上下文获取用户ID，调用服务层领取
// 错误：路径参数非数字返回 400；领取失败返回 500（已领完/达上限等业务错误文案）
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
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetAvailableCouponList 结算可用优惠券接口
// 路由映射：GET /api/v1/users/platform/coupons/available?order_amount=xxx
// 鉴权：JWT（AuthMiddleware 注入用户信息）
// 功能：绑定订单金额参数，从上下文获取用户ID，返回按实付升序的可用券列表
// 错误：参数绑定失败返回 400；order_amount 非法或查询失败返回 500
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
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
