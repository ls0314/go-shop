// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CancelOrder 取消订单(用户端)。
//
// ============================================================
// 这是**反向下的事务**:它要退还已占用的资源
// ============================================================
//
// 下单时做了三件事(Saga 的正向):
//
//	① 锁库存(LockStock)
//	② 核销优惠券(UseCoupon)
//	③ 落订单行(本地事务)
//
// 取消要**逆序**补偿:
//
//	① 释放库存(ReleaseStock)
//	② 退还优惠券(ReturnCoupon)
//	③ 把订单置为已取消
//
// 这一切在 trade-service 内部完成,BFF 只发起。
//
// ============================================================
// 补偿语义:BFF 不重试、不补偿
// ============================================================
//
// 若 trade 侧的补偿失败(如 product-service 抖动导致释放库存失败),
// trade 不会把这个失败透给 BFF —— 它会返回一个业务结果,
// 并由**对账任务**兜底(DS-A-26 的 Saga 设计:补偿失败要能被发现,
// 而不是让用户看到"取消失败,请重试"然后重试又改不了状态)。
//
// 故 BFF 这边**看到的是"取消成功"还是"取消失败"取决于 trade 的
// 判断**,而不是补偿的成败。BFF 不做重试 —— 重试会重复触发补偿,
// 而补偿虽是幂等的(靠 idempotency_key),但每次都多一次 RPC。
//
// **也不要因为"补偿可能失败"就在 BFF 加一个重试循环** ——
// 那是把 trade 的对账职责搬到 BFF,而 BFF 没有那个上下文
// (它不知道哪些订单需要重新对账)。
//
// ============================================================
// 归属校验与"谁能取消"
// ============================================================
//
// proto 的 CancelOrderReq 是 {order_id, user_id, operator},
// 服务端查 `WHERE order_id = ? AND user_id = ?` —— 归属在服务端。
//
// operator 传 JWT 里的 username,用于订单日志("谁取消的")。
// 用户端自己取消,故 operator 就是他自己。
//
// **哪些状态可以取消**由服务端判断(如已发货的不能取消),
// BFF 不预检 —— 预检要先查订单状态(多一次 RPC + TOCTOU),
// 且状态机属于服务端的领域知识。
//
// ============================================================
// compensated 字段:**必须记日志,不能丢**
// ============================================================
//
// proto 对 CancelOrderResp.compensated 的注释:
//
//	"库存释放与券归还是否都已成功。false 表示订单已取消但补偿未完成"
//	"返回里单列 compensated 字段说明补偿是否全部完成,**便于上层告警**"
//
// **BFF 就是那个"上层"**。虽然 HTTP 响应里不返回它(单体没暴露给
// 前端 —— 用户不需要知道补偿细节),但 BFF **必须把它记进日志**:
//
//	compensated=false 意味着订单已取消、但库存没释放或券没退回 ——
//	那是一个需要运维介入的状态。BFF 是这条信息在链路里的**唯一
//	经手方**,不记就等于彻底丢掉。
//
// (trade-service 侧应当也有日志,但 BFF 这一层的记录能定位到
// "哪个用户、哪次请求"触发了它,两边的日志互补。)
//
// 这也是本项目里少数"响应字段不返回给前端但必须处理"的例子 ——
// 若照抄单体的 handler(它只取 order 三个字段),这个信号就丢了。
func (l *CancelOrderLogic) CancelOrder(req *types.OrderIdReq) (*types.CancelOrderResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}
	userName, _ := middleware.Username(l.ctx)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.CancelOrder(ctx, &v1_tradev1.CancelOrderReq{
		OrderId: req.Id,
		UserId:  userId,
		// 订单日志的"操作人"
		Operator: userName,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 订单不存在 / 不属于当前用户 / 状态不允许取消 → 400
		//
		// 前端应展示服务端给的文案 —— 它比"取消失败"更有信息量
		// (如"订单已发货,无法取消")。
		return nil, err
	}

	// 补偿未完成 → 记日志(见上方说明)。用 Error 级别而不是 Warn:
	// 它有明确的运维动作(查库存与券的实际状态并人工收敛)。
	if !resp.GetCompensated() {
		logx.WithContext(l.ctx).Errorf(
			"订单已取消但补偿未完成(需对账收敛): order_id=%d order_no=%s user_id=%d",
			resp.GetOrder().GetOrderId(), resp.GetOrder().GetOrderNo(), userId)
	}

	// resp 里的 order 是内层消息,单体把它摊平到外层
	o := resp.GetOrder()
	return &types.CancelOrderResp{
		OrderId:     o.GetOrderId(),
		OrderNo:     o.GetOrderNo(),
		OrderStatus: o.GetOrderStatus(),
	}, nil
}
