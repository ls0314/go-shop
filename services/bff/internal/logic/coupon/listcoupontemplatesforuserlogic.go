// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package coupon

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCouponTemplatesForUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCouponTemplatesForUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponTemplatesForUserLogic {
	return &ListCouponTemplatesForUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCouponTemplatesForUser 领券中心:当前用户可领取的券模板。
//
// ============================================================
// user_id 从 JWT 取 —— 这条最容易被误判成"不需要身份"
// ============================================================
//
// 它只是"看模板列表",看起来与个人无关;但 proto 的 Req 里 user_id 是
// **必需**的,注释写明用途:
//
//	"user_id 用于算 held_count(该用户已领数量),前端据此置灰'领取'按钮"
//
// 故没有身份就算不出 held_count,列表也就没法正确置灰。取不到返回
// errNoIdentity → 500(该路由未挂 Auth = 配置错误)。
//
// 注意 `GET /coupons` 那条也有 user_id,但用途不同(过滤归属);
// 这里只用来算已领数,模板本身对所有用户是同一份。
//
// ============================================================
// coupon_name / coupon_type 传不下去,只能忽略
// ============================================================
//
// .api 的 UserCouponTemplateListReq 有这两个 form 字段,但 proto 的
// ListCouponTemplatesForUserReq 只有 {user_id, page, page_size} ——
// **服务端没有这个筛选能力**。
//
// 不在 BFF 本地过滤:本地只能筛当前这一页,筛掉几条之后 total 与
// 实际条数就对不上了,分页控件会算错页数。少一个筛选参数比给出错的
// 分页好。
//
// 前端也确实没在用:demo_shop_front 的 GetReceiveCouponListApi 签名
// 只有 {page, page_size}(管理端那个 GetCouponListApi 才带两个筛选)。
// 故这里是"接了但不用",而不是"漏了"。
//
// ============================================================
// 分页兜底并回写
// ============================================================
//
// 与 order 域同口径:不传 → 1/10,page_size 上限 100,兜底后的值回写。
//
// ============================================================
// 每一行是**拍平**的
// ============================================================
//
// proto 给的是 CouponTemplateForUser{template, held_count, remaining_count},
// 而 HTTP 的行是扁平字段(见 convert.go 的 toUserCouponTemplateItems)。
// proto 的 template.status 不回给前端 —— HTTP 契约里没有这个字段。
func (l *ListCouponTemplatesForUserLogic) ListCouponTemplatesForUser(req *types.UserCouponTemplateListReq) (*types.UserCouponTemplateListResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.ListCouponTemplatesForUser(ctx, &v1_marketingv1.ListCouponTemplatesForUserReq{
		UserId:   userId,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.UserCouponTemplateListResp{
		List:     toUserCouponTemplateItems(resp.GetItems()),
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
