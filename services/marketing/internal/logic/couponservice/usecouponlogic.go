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

type UseCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUseCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UseCouponLogic {
	return &UseCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UseCoupon 下单核销券(供 trade-service 调用)。
//
// 步骤与单体下单事务里的那段一致,只是搬到了本服务内:
//  1. 取可用券(不存在/已用/已过期都查不到)
//  2. 校验归属 —— owner 不可变,无竞态
//  3. 条件更新核销,写入幂等键与订单号
//  4. 算实付金额返回
//
// **幂等**:幂等键上有一条部分唯一索引(uk_user_coupon_idem)。重复核销时:
//   - 同一张券:第 3 步的状态条件(status = unused)不匹配 → 影响 0 行 → 返回"券不可用"
//   - 不同券但同键:唯一索引拒绝写入 → 返回基础设施错误,调用方可见
//
// 两种结果都不会让券被用两次 —— 这正是重放该得到的结论。
func (l *UseCouponLogic) UseCoupon(in *v1_marketingv1.UseCouponReq) (*v1_marketingv1.UseCouponResp, error) {
	if in.IdempotencyKey == "" {
		// 幂等键是补偿与重放的唯一凭据,缺了就不能放行。
		// 宁可显式失败,也不要"静默无幂等"地核销
		return &v1_marketingv1.UseCouponResp{ErrorMsg: model.ErrIdempotencyKeyRequired.Error()}, nil
	}

	coupon, err := l.svcCtx.UserCouponRepo.GetUsableCouponById(in.UserCouponId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &v1_marketingv1.UseCouponResp{ErrorMsg: model.ErrCouponNotExistOrUsed.Error()}, nil
		}
		return nil, err
	}

	// 归属校验放在表的所有权方这一侧:调用方可能传错 userId,不能替它兜底
	if coupon.UserId != in.UserId {
		return &v1_marketingv1.UseCouponResp{ErrorMsg: model.ErrUseCouponNoPermission.Error()}, nil
	}

	// 门槛与实付在核销前算 —— 不满足门槛就不该把券用掉
	payAmount, ok := model.CalcPayAmount(coupon.CouponType, in.OrderAmount, coupon.ThresholdAmount, coupon.DiscountAmount)
	if !ok {
		return &v1_marketingv1.UseCouponResp{ErrorMsg: model.ErrCouponThresholdNotMet.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		rows, err := l.svcCtx.UserCouponRepo.WithTx(tx).UseCoupon(in.UserCouponId, in.IdempotencyKey, in.OrderNo)
		if err != nil {
			return err
		}
		if rows == 0 {
			// 并发下被别人先核销了(同一张券不能用在两张单上)
			return model.ErrCouponNotExistOrUsed
		}
		return nil
	})
	if err != nil {
		if isCouponBizError(err) {
			return &v1_marketingv1.UseCouponResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	// 核销后读回权威值:此时券已是 used,必须用不带状态过滤的查法,
	// 否则读不到 —— 返回的快照也就丢掉了 order_no / used_at。
	if updated, err := l.svcCtx.UserCouponRepo.GetUserCouponById(in.UserCouponId); err == nil && updated != nil {
		coupon = updated
	}

	return &v1_marketingv1.UseCouponResp{
		Coupon:    converter.ToProtoUserCoupon(coupon),
		PayAmount: payAmount,
	}, nil
}
