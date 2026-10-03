package service

import (
	"demo-shop-back/src/infra/couponclient"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
)

// ============================================================
//	定义及实例化
// ============================================================

// CouponService 优惠券服务层实例。
//
// 券域(模板 CRUD + 用户券中心 + 领券)**全部已迁 marketing-service**:
// 券表 coupon_template / user_coupon 的所有权在那边,落在独立库 marketing_db。
// 因此本结构不再持有任何 repo,也不再持有闸门缓存 —— 它是纯粹的 RPC 门面。
//
// 领券闸门(coupon:stock / coupon:limit / coupon:ucnt)随表一起迁走了:
// 闸门键的权威值来自券表,留在单体就等于让"读不到表的一方"维护计数器。
type CouponService struct {
	CouponRPC *couponclient.CouponClient
}

// NewCouponService 创建优惠券服务层实例
// 接收值：deps - 服务层依赖（由 composition root 注入）
// 返回值：*CouponService - 优惠券服务层实例指针
func NewCouponService(deps ServiceDeps) *CouponService {
	return &CouponService{
		CouponRPC: deps.CouponRPC,
	}
}

// ============================================================
//	管理端:模板
// ============================================================

// CreateCoupon 创建优惠券模板（接口1）
// 路由映射：POST /api/v1/admin/coupons
// 所需权限：platform:coupon:create
//
// 参数校验链(类型枚举/优惠力度/门槛/总量/限领/有效期二选一)已随业务迁至
// marketing-service —— 它与 coupon_template 的 CHECK 约束是一套规则,
// 拆开维护必然漂移。本地只做 HTTP 绑定与 RPC 转发。
func (c *CouponService) CreateCoupon(req requset.CreateCouponReq) (*response.CreateCouponResp, error) {
	templateId, errMsg, err := c.CouponRPC.CreateCouponTemplate(req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return &response.CreateCouponResp{TemplateId: templateId}, nil
}

// GetCouponList 分页查询优惠券模板列表（接口2）
// 路由映射：GET /api/v1/admin/coupons
// 所需权限：platform:coupon:view
//
// 分页归一化(page<=0→1;pageSize<=0→10;>100 封顶 100)在服务端做,
// 且服务端把**实际生效**的 page/page_size 回显在响应里 —— 本地不再兜一份,
// 避免两边各写一套默认值与封顶值后漂移。
func (c *CouponService) GetCouponList(req requset.GetCouponListReq) (*response.GetCouponListResp, error) {
	resp, errMsg, err := c.CouponRPC.GetCouponList(req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return resp, nil
}

// ============================================================
//	用户端:我的券 / 领券中心 / 结算可用券 / 领取
// ============================================================

// UserGetCouponList 分页查询当前用户的优惠券列表（接口4）
// 路由映射：GET /api/v1/coupons
// 鉴权：JWT（userId 从上下文获取，不支持越权查询他人券）
//
// 与结算可用券的区别:本方法**不过滤过期**,已用/已过期也要展示。
func (c *CouponService) UserGetCouponList(userId int64, req requset.UserGetCouponListReq) (*response.UserGetCouponListResp, error) {
	resp, errMsg, err := c.CouponRPC.GetUserCouponList(userId, req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return resp, nil
}

// receiveCouponRPC 领券的 RPC 实现。
//
// **私有**:对外的 ReceiveCoupon 在 coupon_metrics.go 里 ——
// 那一层包着 Prometheus 埋点(业务失败按文案分类计数),
// 直接调本方法会绕过埋点。
//
// 并发安全的核心实现(双防线:模板行 FOR UPDATE 串行化 + 条件扣减防超发,
// 外加 Redis 闸门挡无效流量)已迁 marketing-service —— 那两道防线的依据
// 是 coupon_template.received_count 与 user_coupon 的行,都在那边。
func (c *CouponService) receiveCouponRPC(userId, templateId int64) (*response.UserReceiveCouponResp, error) {
	resp, errMsg, err := c.CouponRPC.ReceiveCoupon(userId, templateId)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return resp, nil
}

// GetReceiveCouponList 领券中心模板列表（用户端可见可领取的券）
// 路由映射：GET /api/v1/coupons/templates
// 鉴权：JWT（userId 从上下文获取，用于算 held_count）
//
// 空列表是正常业务状态(前端展示"暂无可用券"),不返回错误。
func (c *CouponService) GetReceiveCouponList(userId int64, req requset.UserGetTemplateListReq) (*response.UserCouponTemplateListResp, error) {
	resp, errMsg, err := c.CouponRPC.GetReceiveCouponList(userId, req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return resp, nil
}

// GetAvailableCouponList 结算页可用券（接口5）
// 路由映射：GET /api/v1/coupons/available?order_amount=xxx
// 鉴权：JWT（userId 从上下文获取）
//
// 门槛过滤与"用券后实付金额"的计算都在服务端完成(那是券的语义,
// 不是订单的),列表已按实付升序排好 —— 第一张即"最优券"。
func (c *CouponService) GetAvailableCouponList(userId int64, req requset.GetAvailableCouponReq) (*response.GetAvailableCouponResp, error) {
	resp, errMsg, err := c.CouponRPC.GetAvailableCouponList(userId, req)
	if err != nil {
		return nil, err
	}
	if errMsg != "" {
		return nil, couponclient.RestoreError(errMsg)
	}
	return resp, nil
}
