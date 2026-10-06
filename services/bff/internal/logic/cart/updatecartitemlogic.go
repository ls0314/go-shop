// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package cart

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartItemLogic {
	return &UpdateCartItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateCartItem 更新购物车项的数量与勾选状态。
//
// ============================================================
// **整条更新,不是部分更新** —— 这是本接口最容易踩的坑
// ============================================================
//
// proto 的 UpdateCartItemReq 是 {cart_item_id, user_id, quantity,
// is_selected} —— 四个字段都是**非指针**的,服务端按整条更新处理。
//
// 单体那边是 `c.ShouldBindJSON(&updateCartItem)` 后原样传:
//
//	ci.tradeRPC.UpdateCartItem(id, userId, updateCartItem.Quantity,
//	                           updateCartItem.IsSelected)
//
// 它**不区分"没传"与"传了零值"**。所以:
//
//	前端只传 {"quantity": 3}
//	  → is_selected 绑定为 false(零值)
//	  → RPC 收到 is_selected=false
//	  → **把已勾选的商品取消勾选**
//
// 故前端改数量时**必须把 is_selected 一起传**(通常是它当前的值)。
//
// **BFF 不"修正"这个行为** —— 比如"只有 quantity 变化就保留原
// is_selected"。那需要先查一次购物车拿到当前状态,把一次写变成
// "读+写",两个并发请求会互相覆盖。且那会改变既有的前端契约
// (前端现在就是两个字段都传的)。
//
// 这也是 .api 里两个字段都不标 optional 的原因 —— 标了会让实现者
// 以为要区分"没传",从而发明一套单体与前端都没有的语义。
//
// ============================================================
// 归属校验在服务端
// ============================================================
//
// proto 注释:"归属人。服务端校验 cart_item.user_id 必须等于它 ——
// 防越权改他人购物车"。
//
// 故 BFF 只传 JWT 里的 userId,不做预检(预检是 TOCTOU 且多一次 RPC)。
// 传别人的 cart_item_id 时服务端会返回"无权操作",BFF 回 400。
//
// ============================================================
// 响应是**裸 CartItem**
// ============================================================
//
// 单体: utils.Success(c, resp),resp 是 *response.CartItemListResp。
// 故 data 就是一个购物车项对象(含更新后的商品侧回填字段),
// 前端可以用它直接刷新那一行,不必重拉列表。
func (l *UpdateCartItemLogic) UpdateCartItem(req *types.UpdateCartItemReq) (*types.CartItem, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.UpdateCartItem(ctx, &v1_tradev1.UpdateCartItemReq{
		CartItemId: req.Id,
		// 归属:服务端用它做 WHERE 校验(见上方说明)
		UserId: userId,
		// 两个字段都原样传 —— 不区分"没传"(见上方说明)
		Quantity:   req.Quantity,
		IsSelected: req.IsSelected,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 库存不足 / 购物车项不存在 / 无权操作 → 400
		return nil, err
	}

	item := toCartItem(resp.GetItem())
	return &item, nil
}
