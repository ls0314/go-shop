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

type AddCartItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCartItemLogic {
	return &AddCartItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddCartItem 加入购物车。
//
// ============================================================
// 响应被**降级**了 —— 这是单体遗留的形状
// ============================================================
//
// proto 的 CreateCartItemResp 返回**整行 CartItem**,注释写明:
//
//	"返回写入后的整行(含商品侧回填字段),前端不必再拉一次列表。"
//
// 但单体的 handler 把它压成了两个字段:
//
//	resp := response.CartItemCreateResp{
//		CartItemId: cartItemId,
//		Quantity:   quantity,
//	}
//
// 即 proto 给的整行在 HTTP 层被丢掉了,前端只拿到 {cart_item_id, quantity}。
//
// **BFF 保持这个降级**(前端零感知)。
//
// 要升回整行是一次**契约增强**而非破坏性变更(只是多给字段,
// 前端现有的 {cart_item_id, quantity} 取值不变),而且能省掉前端
// 加购后的一次列表请求。但那是独立的一次改动,应当单独做并
// 通知前端 —— 不要夹在迁移里顺手改。
//
// ============================================================
// user_id 从 JWT 取
// ============================================================
//
// .api 里 AddCartItemReq 只有 {sku_id, quantity} —— 没有 user_id。
// 这是刻意的:归属必须由服务端从令牌决定,否则任何登录用户都能往
// 别人的购物车里塞东西。
//
// ============================================================
// 归属校验在服务端
// ============================================================
//
// 这条是"新增",本来就不存在越权(它建的是自己的行)。
// 但同组的 Update / Delete 会传 user_id 给服务端做
// `WHERE cart_item_id = ? AND user_id = ?` —— 见各自的说明。
func (l *AddCartItemLogic) AddCartItem(req *types.AddCartItemReq) (*types.AddCartItemResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.CreateCartItem(ctx, &v1_tradev1.CreateCartItemReq{
		UserId:   userId,
		SkuId:    req.SkuId,
		Quantity: req.Quantity,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 库存不足 / 商品已下架 / SKU 不存在 → 400
		return nil, err
	}

	// 降级成两个字段(见上方说明)。
	//
	// 取值方式与单体一致(已核实 tradeclient.CreateCartItem 用的是
	// `resp.GetItem().GetCartItemId()`)。故若 trade-service 没填 item,
	// 两边都会得到 cart_item_id=0 —— **不是 BFF 引入的差异**,
	// 但值得在端到端验证时确认一次:加购后响应里的 cart_item_id 非 0。
	item := resp.GetItem()
	return &types.AddCartItemResp{
		CartItemId: item.GetCartItemId(),
		Quantity:   item.GetQuantity(),
	}, nil
}
