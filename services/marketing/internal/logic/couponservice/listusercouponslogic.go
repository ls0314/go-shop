package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/converter"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserCouponsLogic {
	return &ListUserCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListUserCoupons 我的卡券分页列表(不过滤过期,已用/已过期也要展示)。
// status 为空表示不过滤。
func (l *ListUserCouponsLogic) ListUserCoupons(in *v1_marketingv1.ListUserCouponsReq) (*v1_marketingv1.ListUserCouponsResp, error) {
	page, pageSize := normalizePage(int(in.Page), int(in.PageSize))

	coupons, total, err := l.svcCtx.UserCouponRepo.GetUserCouponList(in.UserId, page, pageSize, in.Status)
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.ListUserCouponsResp{
		Items:    converter.ToProtoUserCouponList(coupons),
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
