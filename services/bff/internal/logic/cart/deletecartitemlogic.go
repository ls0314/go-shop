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

type DeleteCartItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCartItemLogic {
	return &DeleteCartItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteCartItem 从购物车删除一项。
//
// ============================================================
// 归属校验在服务端
// ============================================================
//
// proto 的 DeleteCartItemReq 是 {cart_item_id, user_id},服务端的查询
// 带着 `WHERE cart_item_id = ? AND user_id = ?` —— 删别人的项会
// 查不到并返回业务错误。
//
// 故 BFF 只传 JWT 里的 userId,不预检。
//
// ============================================================
// 响应:单体是 null,这里是 {}
// ============================================================
//
// 单体: utils.Success(c, nil) → "data": null
// BFF:  returns (Empty) → &types.Empty{} → "data": {}
//
// **这是全项目 14 条 returns (Empty) 路由的共同差异**,已在多处标注。
// 若前端判 `res.data === null` 会走错分支 —— 端到端验证时确认。
//
// 幂等:删一个不存在的项,服务端返回什么由它决定(BFF 透传)。
// 若它报"购物车项不存在"→ 400;若它静默成功 → 200。
// 前端的"删除"按钮通常不关心区分这两种。
func (l *DeleteCartItemLogic) DeleteCartItem(req *types.CartItemIdReq) (*types.Empty, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.DeleteCartItem(ctx, &v1_tradev1.DeleteCartItemReq{
		CartItemId: req.Id,
		UserId:     userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 购物车项不存在 / 无权操作 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
