package couponservicelogic

import (
	"context"
	"sort"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/model"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAvailableCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableCouponsLogic {
	return &ListAvailableCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAvailableCoupons 结算页可用券:未过期 + unused,且满足订单金额门槛。
// pay_after 由服务端算(满减/直减的差别不外泄),列表按 pay_after 升序 ——
// 第一张即"最优券"。
func (l *ListAvailableCouponsLogic) ListAvailableCoupons(in *v1_marketingv1.ListAvailableCouponsReq) (*v1_marketingv1.ListAvailableCouponsResp, error) {
	if in.OrderAmount < 0 {
		return &v1_marketingv1.ListAvailableCouponsResp{ErrorMsg: model.ErrCouponOrderAmountInvalid.Error()}, nil
	}

	coupons, err := l.svcCtx.UserCouponRepo.GetAvailableCouponList(in.UserId)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_marketingv1.AvailableCoupon, 0, len(coupons))
	for _, c := range coupons {
		payAfter, ok := model.CalcPayAmount(c.CouponType, in.OrderAmount, c.ThresholdAmount, c.DiscountAmount)
		if !ok {
			// 未达门槛的券不进结算列表 —— 与单体一致(它在循环里 continue)
			continue
		}
		items = append(items, &v1_marketingv1.AvailableCoupon{
			UserCouponId:    c.UserCouponId,
			CouponName:      c.CouponName,
			CouponType:      c.CouponType,
			ThresholdAmount: c.ThresholdAmount,
			DiscountAmount:  c.DiscountAmount,
			PayAfter:        payAfter,
		})
	}

	// 按实付金额升序:排在最前的即"最优券",由前端标记展示
	sort.Slice(items, func(i, j int) bool {
		return items[i].PayAfter < items[j].PayAfter
	})

	return &v1_marketingv1.ListAvailableCouponsResp{Items: items}, nil
}
