package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/model"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteCouponTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCouponTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCouponTemplateLogic {
	return &DeleteCouponTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteCouponTemplate 软删除模板(置 is_deleted)。
//
// 不物理删除:已领出的 user_coupon 仍指向它,硬删会让用户券列表丢掉名称与规则。
// 软删后模板不再出现在管理端与领券中心,已领的券照常可用。
func (l *DeleteCouponTemplateLogic) DeleteCouponTemplate(in *v1_marketingv1.DeleteCouponTemplateReq) (*v1_marketingv1.DeleteCouponTemplateResp, error) {
	_, err := l.svcCtx.CouponRepo.GetCouponById(in.TemplateId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &v1_marketingv1.DeleteCouponTemplateResp{ErrorMsg: model.ErrCouponTemplateNotExist.Error()}, nil
		}
		return nil, err
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.CouponRepo.WithTx(tx).DeleteCoupon(in.TemplateId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.DeleteCouponTemplateResp{}, nil
}
