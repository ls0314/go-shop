package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/converter"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCouponTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCouponTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponTemplatesLogic {
	return &ListCouponTemplatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListCouponTemplates 管理端分页查询模板。coupon_name 精确匹配、coupon_type 枚举筛选。
func (l *ListCouponTemplatesLogic) ListCouponTemplates(in *v1_marketingv1.ListCouponTemplatesReq) (*v1_marketingv1.ListCouponTemplatesResp, error) {
	page, pageSize := normalizePage(int(in.Page), int(in.PageSize))

	coupons, total, err := l.svcCtx.CouponRepo.GetCouponList(page, pageSize, in.CouponName, in.CouponType)
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.ListCouponTemplatesResp{
		Items:    converter.ToProtoCouponTemplateList(coupons),
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
