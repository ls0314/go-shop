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

type ReceiveCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReceiveCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReceiveCouponLogic {
	return &ReceiveCouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReceiveCoupon 领取优惠券。
//
// ============================================================
// 路径参数 :id 是 template_id,不是 user_coupon_id
// ============================================================
//
// 领券时用户券还不存在(id 由服务端领券时生成),故 :id 只能是模板 id。
// 路由 POST /coupons/receive/:id 与单体一致
// (POST /api/v1/users/platform/coupons/receive/:id)。
//
// ============================================================
// user_id 从 JWT 取
// ============================================================
//
// proto 的 ReceiveCouponReq 是 {user_id, template_id},HTTP 请求里只有
// 路径参数。归属必须由服务端从令牌决定 —— 否则任何人都能给**别人**
// 领券(把别人的人头数刷满 per_user_limit)。
//
// 服务端还会按 user_id 做限领校验(每人限领数)与总量扣减,故这里的
// userId 同时是"领给谁"与"算谁的名额"。
//
// 这条路由单独挂了 PublicRateLimit(见 bff.api):领券是写操作且可被
// 脚本刷,单体也在这一条上挂了 PerIPRateLimit。
//
// ============================================================
// expire_time 与 expire_at:同一个概念的两个名字
// ============================================================
//
// proto 的字段名是 **expire_at**(ReceiveCouponResp 的 2 号字段,
// getter 是 GetExpireAt),而 HTTP 契约里叫 **expire_time**
// (types.ReceiveCouponResp / 单体 response.UserReceiveCouponResp)。
//
// 同一个 proto 字段,在"我的券"列表里叫 expire_at、在领券响应里叫
// expire_time —— 这是单体时代就有的不一致,**照实实现,不要统一**:
// 前端读的就是 expire_time,改名会让领券成功后拿不到有效期。
//
// 取不到 expire_at 时为 nil,格式化成空串(见 formatTimestamp)。
//
// ============================================================
// 失败都经 error_msg 回来
// ============================================================
//
// 模板不存在(11001)/ 已领完(11002)/ 超过每人限领(11003)都是
// 业务失败 → Classify 落 ResultBiz → 400,前端按 message 提示。
func (l *ReceiveCouponLogic) ReceiveCoupon(req *types.ReceiveCouponReq) (*types.ReceiveCouponResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.ReceiveCoupon(ctx, &v1_marketingv1.ReceiveCouponReq{
		UserId:     userId,
		TemplateId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ReceiveCouponResp{
		UserCouponId: resp.GetUserCouponId(),
		// 改名:proto 的 expire_at → HTTP 的 expire_time(见上方说明)。
		ExpireTime: formatTimestamp(resp.GetExpireAt()),
	}, nil
}
