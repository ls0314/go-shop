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

type UpdateCouponTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponTemplateLogic {
	return &UpdateCouponTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCouponTemplate 局部更新模板,返回合并后的完整对象。
//
// 已发放的模板只允许改名称/有效期这类展示字段:门槛与优惠力度一旦被改,
// 已领出的券会按新规则结算,对已下单的用户不公平。校验放在这里而不是 DB。
func (l *UpdateCouponTemplateLogic) UpdateCouponTemplate(in *v1_marketingv1.UpdateCouponTemplateReq) (*v1_marketingv1.UpdateCouponTemplateResp, error) {
	old, err := l.svcCtx.CouponRepo.GetCouponById(in.TemplateId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &v1_marketingv1.UpdateCouponTemplateResp{ErrorMsg: model.ErrCouponTemplateNotExist.Error()}, nil
		}
		return nil, err
	}

	newCoupon := *old
	if err := decodeUpdates(&newCoupon, in.Updates); err != nil {
		return &v1_marketingv1.UpdateCouponTemplateResp{ErrorMsg: err.Error()}, nil
	}

	// 已发放过就不许动计价相关的三要素与总量
	if old.ReceivedCount > 0 {
		if newCoupon.ThresholdAmount != old.ThresholdAmount ||
			newCoupon.DiscountAmount != old.DiscountAmount ||
			newCoupon.CouponType != old.CouponType ||
			newCoupon.TotalCount != old.TotalCount {
			return &v1_marketingv1.UpdateCouponTemplateResp{ErrorMsg: model.ErrCouponParamInvalid.Error()}, nil
		}
	}

	// 改动后的有效期仍须合法(只改其中一个字段也可能破坏二选一)
	if err := validateTemplate(&newCoupon); err != nil {
		if isCouponBizError(err) {
			return &v1_marketingv1.UpdateCouponTemplateResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.CouponRepo.WithTx(tx).UpdateCoupon(&newCoupon)
	})
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.UpdateCouponTemplateResp{
		Template: converter.ToProtoCouponTemplate(&newCoupon),
	}, nil
}
