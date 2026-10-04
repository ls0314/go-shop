package couponservicelogic

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/model"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ReturnCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReturnCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReturnCouponLogic {
	return &ReturnCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReturnCoupon 取消订单归还券(供 trade-service 的 Saga 补偿调用)。
//
// **按 idempotency_key 反查,而不是 order_no**:补偿会被重放,而 order_no
// 在第一次归还时就被清空了 —— 用它反查第二次会落空,进而把"已经取消成功的
// 订单"报成失败。幂等键虽然也在归还时清空,但它与"券回到 unused"是
// **同一次更新**,所以重放时查不到恰恰说明已经归还完,是幂等命中而非错误。
//
// 三种情况都返回成功(returned 区分):
//   - 键能查到且券是 used → 真归还,returned = true
//   - 键能查到但券已是 unused → 重复补偿,returned = false
//   - 键查不到 → 该单没用券,或已归还过,returned = false
//
// 补偿必须比业务动作更宽容:否则重复投递的取消消息会把已经取消成功的
// 订单报成失败。这与单体不同 —— 单体在取消事务里归还失败会让整个取消失败。
func (l *ReturnCouponLogic) ReturnCoupon(in *v1_marketingv1.ReturnCouponReq) (*v1_marketingv1.ReturnCouponResp, error) {
	if in.IdempotencyKey == "" {
		return &v1_marketingv1.ReturnCouponResp{ErrorMsg: model.ErrIdempotencyKeyRequired.Error()}, nil
	}

	coupon, err := l.svcCtx.UserCouponRepo.GetUserCouponByIdempotencyKey(in.IdempotencyKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 该单没用券,或已经归还过(键被清空)→ 补偿目标已达成
			return &v1_marketingv1.ReturnCouponResp{Returned: false}, nil
		}
		return nil, err
	}

	// userId 为 0 表示不校验:系统取消订单没有操作人语义
	if in.UserId != 0 && coupon.UserId != in.UserId {
		return &v1_marketingv1.ReturnCouponResp{ErrorMsg: model.ErrUseCouponNoPermission.Error()}, nil
	}

	var returned bool
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		rows, err := l.svcCtx.UserCouponRepo.WithTx(tx).ReturnCoupon(coupon.UserCouponId)
		if err != nil {
			return err
		}
		// rows == 0:券不是 used(已被归还 / 已过期)→ 幂等命中,不算失败
		returned = rows > 0
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.ReturnCouponResp{Returned: returned}, nil
}
