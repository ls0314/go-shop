package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/converter"
	"demo-shop/services/marketing/internal/model"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetCouponTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCouponTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCouponTemplateLogic {
	return &GetCouponTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCouponTemplate 按ID查模板
func (l *GetCouponTemplateLogic) GetCouponTemplate(in *v1_marketingv1.GetCouponTemplateReq) (*v1_marketingv1.GetCouponTemplateResp, error) {
	coupon, err := l.svcCtx.CouponRepo.GetCouponById(in.TemplateId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &v1_marketingv1.GetCouponTemplateResp{ErrorMsg: model.ErrCouponTemplateNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_marketingv1.GetCouponTemplateResp{
		Template: converter.ToProtoCouponTemplate(coupon),
	}, nil
}
