package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/converter"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCouponTemplatesForUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCouponTemplatesForUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponTemplatesForUserLogic {
	return &ListCouponTemplatesForUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListCouponTemplatesForUser 领券中心:可领模板 + 该用户已领数/剩余量。
// 前端据 held_count >= per_user_limit 置灰"领取"按钮。
func (l *ListCouponTemplatesForUserLogic) ListCouponTemplatesForUser(in *v1_marketingv1.ListCouponTemplatesForUserReq) (*v1_marketingv1.ListCouponTemplatesForUserResp, error) {
	page, pageSize := normalizePage(int(in.Page), int(in.PageSize))

	coupons, total, err := l.svcCtx.CouponRepo.ListCouponForUser(in.UserId, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.ListCouponTemplatesForUserResp{
		Items:    converter.ToProtoCouponForUserList(coupons),
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
