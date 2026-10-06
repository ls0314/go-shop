// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateOrder 下单。**这是全项目唯一一个两跳的接口**。
//
// ============================================================
// 两跳的顺序不能颠倒
// ============================================================
//
//	① user-service 的 GetAddressSnapshot(userId, addressId)
//	② trade-service 的 CreateOrder(含地址快照)
//
// 为什么不让 trade 拿 address_id 自己查(proto 注释也写了这条):
//
//	① 地址表在 user_db,trade 跨库取不到;
//	② 订单必须存"下单那一刻"的地址 —— 用户随后改地址不能影响
//	   历史订单,所以传快照比传 id 更贴合语义。
//
// 故 .api 里 CreateOrderReq 有 address_id,但它**不传给 trade** ——
// 只用于第一跳。
//
// ============================================================
// 为什么用 GetAddressSnapshot 而不是 GetAddress
// ============================================================
//
// 这是 DS-A-25 §4.5.2 第 1 条专门点的:两者的**失败语义不同**。
//
//	GetAddressSnapshot
//	  "地址不存在"  → ErrorMsg(业务失败,不该重试)
//	  user-service 挂 → gRPC error(该重试)
//
//	GetAddress
//	  把两种都还原成 ErrUnavailable
//
// 下单时把"地址不存在的"报成"服务不可用",会给用户一个
// **误导性提示**("系统繁忙,请稍后重试"),而他实际要做的是
// 重新选一个地址。故必须用前者。
//
// user-service 侧的实现也印证了这一点(它的注释):
//
//	"走与其它操作同一个归属校验 —— 下单取快照同样是'读自己的地址',
//	 不校验就等于允许拿别人的地址下单。"
//
// 所以归属校验在**服务端**做,BFF 只传 JWT 里的 userId。
// 传别人的 address_id 会得到"地址不存在"(服务端不区分
// "不存在"与"不属于你",这是刻意的,避免泄露他人地址的存在性)。
//
// ============================================================
// 幂等键由前端生成
// ============================================================
//
// create_order_req.idempotent_key 由**前端生成**(UUID),全局唯一 ——
// 它是所有跨服务操作的幂等判据,**不是 order_no**。
//
// 用户手抖点两次"提交订单"时,两次请求带同一个 key,
// trade-service 靠它返回同一个订单而不是建两个。
//
// **BFF 不生成也不修改它** —— 生成就等于每次重试都是新订单,
// 幂等彻底失效。若前端没传(空串),trade-service 的行为由它决定
// (可能退化成"每次新建")。
func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (*types.CreateOrderResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}
	userName, _ := middleware.Username(l.ctx)

	// ---------- 第 ① 跳:取地址快照 ----------
	addrCtx, addrCancel := rpc.CtxWithTimeout(l.ctx)
	defer addrCancel()

	addrResp, addrGrpcErr := l.svcCtx.AddressRPC.GetAddressSnapshot(addrCtx, &v1_userv1.GetAddressSnapshotReq{
		UserId:    userId,
		AddressId: req.AddressId,
	})

	addrKind, addrErr := rpc.Classify(addrResp.GetErrorMsg(), addrGrpcErr)
	switch addrKind {
	case rpc.ResultInfra:
		// user-service 挂了 → 503。**这与"地址不存在"必须区分开**,
		// 见上方 GetAddressSnapshot vs GetAddress 的说明。
		return nil, addrErr
	case rpc.ResultBiz:
		// 地址不存在 / 不属于当前用户 → 400,文案由服务端给
		// (如"地址不存在")。前端据此提示用户重新选地址。
		return nil, addrErr
	}

	// ---------- 第 ② 跳:下单 ----------
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.CreateOrder(ctx, &v1_tradev1.CreateOrderReq{
		UserId: userId,
		// 用户名快照:trade 把它写进订单,历史订单不随改名的用户而变
		Username: userName,
		// 字段名是 Address(proto: `AddressSnapshot address = 3`)
		Address:       toTradeSnapshot(addrResp.GetSnapshot()),
		IdempotentKey: req.IdempotentKey,
		BuyerRemark:   req.BuyerRemark,
		UserCouponId:  req.UserCouponId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 库存不足 / 券不可用 / 购物车为空 / 幂等键冲突 → 400
		//
		// **注意**:Saga 的补偿结果不体现在这里 —— 若 trade 侧
		// 补偿成功,它返回的是业务失败(400);补偿失败会由它的
		// 对账任务处理,不通过这个接口暴露(DS-A-26 的 Saga 设计)。
		return nil, err
	}

	// 响应字段与单体 response.CreateOrderResp 一致
	// (order_id / order_no / pay_amount / total_amount /
	//  order_status / pay_expire_at / created_at)。
	return &types.CreateOrderResp{
		OrderId:     resp.GetOrderId(),
		OrderNo:     resp.GetOrderNo(),
		PayAmount:   resp.GetPayAmount(),
		TotalAmount: resp.GetTotalAmount(),
		OrderStatus: resp.GetOrderStatus(),
		PayExpireAt: formatTimestamp(resp.GetPayExpireAt()),
		CreatedAt:   formatTimestamp(resp.GetCreatedAt()),
	}, nil
}
