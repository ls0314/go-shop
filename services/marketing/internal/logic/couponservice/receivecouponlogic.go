package couponservicelogic

import (
	"context"
	"time"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/infra/gate"
	"demo-shop/services/marketing/internal/model"
	"demo-shop/services/marketing/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type ReceiveCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReceiveCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReceiveCouponLogic {
	return &ReceiveCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReceiveCoupon 领券。
//
// 并发设计(双防线,与单体一致):
//
//	① 悲观锁:SELECT ... FOR UPDATE 锁模板行,同一模板的并发领取串行排队;
//	   锁内校验限领(CountUserCoupon)与插入成为原子操作 —— 防每人超领
//	② 乐观锁:UPDATE ... WHERE received_count < total_count 条件扣减,
//	   影响行数 0 即售罄 —— 防总量超发
//
// 闸门(Redis)是可选的第三层:它只挡无效流量,**正确性不依赖它** ——
// Redis 不可用时整条闸门旁路,由上面两道防线独立保证不超发。
//
// 埋点见本函数末尾:结果分类在此处才准(sold_out/limit_exceeded 可能来自
// 闸门也可能来自 DB 双防线,业务语义相同),path 标签也只有在**本进程内**
// 才能填出真值 —— 详见 infra/metrics 的说明。
func (l *ReceiveCouponLogic) ReceiveCoupon(in *v1_marketingv1.ReceiveCouponReq) (*v1_marketingv1.ReceiveCouponResp, error) {
	start := time.Now()
	// gateServed 记录"闸门是否真的做出了判定"。
	//
	// 与"闸门是否启用"不是一回事:启用了但本次异常降级、或回填失败后
	// 仍落到 DB,都算 db_only。这正是单体那侧填不出来的区分。
	gateServed := false

	gatePassed := false
	g := l.svcCtx.CouponGate
	if g.Enabled() {
		code, err := g.Deduct(l.ctx, in.TemplateId, in.UserId)
		if err != nil {
			// 闸门异常不阻断领取:降级直走 DB。闸门是加速器,不是正确性来源
			l.Errorf("领券闸门异常,降级直走 DB: templateId=%d err=%v", in.TemplateId, err)
		} else {
			if code == gate.GateBackfill {
				if l.backfillGate(in.TemplateId) {
					code, err = g.Deduct(l.ctx, in.TemplateId, in.UserId)
				}
			}
			if err == nil {
				switch {
				case code > 0:
					gatePassed = true
					gateServed = true
				case code == gate.GateSoldOut:
					// 闸门就挡下了:这是**闸门生效**的证据(而非 DB 双防线挡的)
					observeReceive("sold_out", true, start)
					return &v1_marketingv1.ReceiveCouponResp{ErrorMsg: model.ErrCouponSoldOut.Error()}, nil
				case code == gate.GateLimitExceeded:
					observeReceive("limit_exceeded", true, start)
					return &v1_marketingv1.ReceiveCouponResp{ErrorMsg: model.ErrCouponLimitExceeded.Error()}, nil
				}
				// code 仍为 GateBackfill(回填失败)→ 落到下方照常走 DB(降级语义)
			}
		}
	}

	var resp *v1_marketingv1.ReceiveCouponResp
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		couponTx := l.svcCtx.CouponRepo.WithTx(tx)
		userCouponTx := l.svcCtx.UserCouponRepo.WithTx(tx)

		// 锁定模板行 —— 从这里开始,同一模板的并发领取在此排队
		tpl, err := couponTx.LockCouponById(in.TemplateId)
		if err != nil {
			return err // 模板不存在时透传 gorm.ErrRecordNotFound
		}

		// 锁内校验(此刻其他领取事务都在等锁,读到的必是最新数据)
		if tpl.IsDeleted {
			return model.ErrCouponTemplateNotExist
		}
		if tpl.ReceivedCount >= tpl.TotalCount {
			return model.ErrCouponSoldOut
		}
		held, err := userCouponTx.CountUserCoupon(in.UserId, in.TemplateId)
		if err != nil {
			return err
		}
		if held >= tpl.PerUserLimit {
			return model.ErrCouponLimitExceeded
		}

		// 原子条件扣减(行锁 + 条件更新双保险)
		rowsAffected, err := couponTx.IncrReceivedCount(in.TemplateId, tpl.TotalCount)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return model.ErrCouponSoldOut
		}

		// 过期时间与插入:数据全部来自锁内读到的 tpl
		expireAt := calcExpireAt(tpl, time.Now())
		coupon := &model.UserCoupon{
			TemplateId: in.TemplateId,
			UserId:     in.UserId,
			ExpireAt:   expireAt,
			Status:     model.CouponUnused,
		}
		if err := userCouponTx.CreateUserCoupon(coupon); err != nil {
			return err
		}

		resp = &v1_marketingv1.ReceiveCouponResp{
			UserCouponId: coupon.UserCouponId,
			ExpireAt:     timestamppb.New(expireAt),
		}
		return nil // commit:锁释放,下一个排队的领取事务开始执行
	})

	// 闸门补偿:DB 失败说明本次领取并未发生,把闸门扣减的还回去
	if err != nil {
		if gatePassed && g.Enabled() {
			if cErr := g.Compensate(context.Background(), in.TemplateId, in.UserId); cErr != nil {
				l.Errorf("领券闸门补偿失败(等待对账收敛): templateId=%d err=%v", in.TemplateId, cErr)
			}
		}
		// 业务失败 → error_msg;基础设施故障 → gRPC error
		if isCouponBizError(err) || err == gorm.ErrRecordNotFound {
			observeReceive(receiveResultOf(err), gateServed, start)
			return &v1_marketingv1.ReceiveCouponResp{ErrorMsg: couponErrText(err)}, nil
		}
		// 基础设施故障:归入 error。这里**不能**当成业务失败 ——
		// 上游据此重试,而业务失败重试没有意义
		observeReceive("error", gateServed, start)
		return nil, err
	}
	observeReceive("success", gateServed, start)
	return resp, nil
}

// backfillGate 从 DB 读权威值回填闸门计数器。
// 回填失败返回 false,调用方降级直走 DB —— 不阻断领取。
func (l *ReceiveCouponLogic) backfillGate(templateId int64) bool {
	tpl, err := l.svcCtx.CouponRepo.GetCouponById(templateId)
	if err != nil {
		return false
	}
	remaining := tpl.TotalCount - tpl.ReceivedCount
	if remaining < 0 {
		remaining = 0
	}
	ttl := gate.GateTTL(tpl)
	l.svcCtx.CouponGate.Fill(l.ctx, gate.StockKey(templateId), remaining, ttl)
	l.svcCtx.CouponGate.Fill(l.ctx, gate.LimitKey(templateId), tpl.PerUserLimit, ttl)
	return true
}
