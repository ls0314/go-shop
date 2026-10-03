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
// 按 order_no 反查而不是让调用方传 user_coupon_id:取消链路上调用方手里
// 只有订单号,而且"该订单用了哪张券"是券域的账 —— 让调用方自己记住再回传,
// 等于把归属信息复制一份出去,必然不同步。
//
// 幂等与"没用券"都是成功:
//   - 券已是 unused(重复补偿)→ 条件更新影响 0 行,返回 returned=false 且无错误
//   - 该订单压根没用券 → 反查不到,同样算成功(补偿的目标已达成)
//
// 这与单体不同:单体在取消事务里 RefundCoupon 返回 0 行会报 ErrCannotCancelCoupon
// 让整个取消失败。拆出后补偿独立成一步,再那样做会让"重复投递的取消消息"
// 把已经取消成功的订单报成失败 —— 补偿动作必须比业务动作更宽容。
func (l *ReturnCouponLogic) ReturnCoupon(in *v1_marketingv1.ReturnCouponReq) (*v1_marketingv1.ReturnCouponResp, error) {
	coupon, err := l.svcCtx.UserCouponRepo.GetUserCouponByOrderNo(in.OrderNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 该订单没用券(或券已被归还并清空了 order_no)→ 目标已达成
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
