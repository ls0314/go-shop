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

type ListAvailableCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableCouponsLogic {
	return &ListAvailableCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListAvailableCoupons 结算页可用券:当前用户在这个订单金额下能用的券。
//
// ============================================================
// user_id 从 JWT 取
// ============================================================
//
// proto 的 Req 是 {user_id, order_amount},而 .api 的 AvailableCouponReq
// 只有 order_amount(form,optional)—— 人的身份不能由前端指定,否则
// 谁都能拿别人的券去试算。取不到返回 errNoIdentity → 500。
//
// ============================================================
// 响应**只有 list**,没有 total / page
// ============================================================
//
// 本条不是分页接口:proto 的 ListAvailableCouponsResp 只有 items +
// error_msg,HTTP 也只有 {list: [...]}。前端拿整个数组。
//
// 不让它分页是有意义的:列表已按实付升序排好,第一张就是"最优券";
// 分页会把"最优"散到不同页去,而且结算页展示的是"这个订单能用哪些券",
// 一个用户的可用券天然有限。
//
// ============================================================
// pay_after 是服务端算的,**BFF 绝不重算**
// ============================================================
//
// 它表示"用这张券之后的应付金额"。算法在券域
// (model.CalcPayAmount):
//
//	full_reduction   → 金额 - 优惠
//	direct_discount  → 金额 × 折扣率
//	未达门槛         → 该券根本不进列表
//
// 券域自己的注释写明"放在券域而不是调用方:这是券的语义,不是订单的",
// 单体时代下单链路里就复制过一份同样的算法。BFF 若再算一遍,一旦优惠
// 语义变化(满减封顶、折扣率取整、金额下限 0 之类),两份实现必然漂移
// ——而结算页显示 470、下单却扣 480 是这类漂移最典型的后果。
// 故这里只做字段搬运,连排序都不做(服务端已按 pay_after 升序)。
//
// ============================================================
// order_amount 直传,不兜默认值
// ============================================================
//
// optional float,不传时为 0。0 是**合法入参**(表示"还不知道金额",
// 服务端按门槛 0 过滤,只有无门槛券会返回),故不能替它编一个默认值 ——
// 编出来的金额会让页面显示一个用户当前订单并不成立的实付金额。
//
// 负数由服务端拒(error_msg 11009 → 400),BFF 不重复校验。
func (l *ListAvailableCouponsLogic) ListAvailableCoupons(req *types.AvailableCouponReq) (*types.AvailableCouponResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.ListAvailableCoupons(ctx, &v1_marketingv1.ListAvailableCouponsReq{
		UserId:      userId,
		OrderAmount: req.OrderAmount,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.AvailableCouponResp{
		List: toAvailableCouponItems(resp.GetItems()),
	}, nil
}
