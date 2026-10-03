package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/converter"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateCouponTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCouponTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCouponTemplateLogic {
	return &CreateCouponTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateCouponTemplate 创建优惠券模板。
// 参数校验链见 validateTemplate;received_count 走数据库默认值 0,后续领取时原子扣减。
func (l *CreateCouponTemplateLogic) CreateCouponTemplate(in *v1_marketingv1.CreateCouponTemplateReq) (*v1_marketingv1.CreateCouponTemplateResp, error) {
	tpl := converter.FromProtoCouponTemplate(in.Template)

	if err := validateTemplate(tpl); err != nil {
		if isCouponBizError(err) {
			return &v1_marketingv1.CreateCouponTemplateResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.CouponRepo.WithTx(tx).CreateCoupon(tpl)
	})
	if err != nil {
		return nil, err
	}

	// 返回完整实体(含服务端生成的自增 id 与推导出的 status),与 CreateScope 同风格
	return &v1_marketingv1.CreateCouponTemplateResp{
		Template: converter.ToProtoCouponTemplate(tpl),
	}, nil
}
