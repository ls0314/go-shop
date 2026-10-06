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

type SelectAllCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSelectAllCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SelectAllCartLogic {
	return &SelectAllCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SelectAllCart 全选/取消全选当前用户的购物车。
//
// ============================================================
// 请求体只有一个字段:is_selected
// ============================================================
//
// .api 里 SelectAllCartReq 只有 {is_selected} —— **没有 user_id**
// (从 JWT 取),也**没有 cart_item_ids**(是全选,不是批量选)。
//
// 后者值得注意:这个接口作用于**该用户的全部购物车项**,
// 不接受 ID 列表。若前端想做"批量勾选选中的几项",那个需求
// 这条接口满足不了 —— 它要么逐条调 UpdateCartItem,
// 要么需要一个新接口。**不要试图用这条接口表达批量选择**。
//
// proto 注释只写了"全选",故语义是从"该用户的所有行"这个角度定义的。
//
// ============================================================
// 响应是 {affected: N}
// ============================================================
//
// proto 的 SelectAllCartItemsResp 是 {affected, error_msg},
// 而单体手工包了一层:
//
//	utils.Success(c, gin.H{"affected": total})
//
// 形状恰好一致(都是 {affected: N}),故 .api 里直接用
// types.SelectAllCartResp{Affected}。
//
// **但那层 gin.H 是有意义的**:它说明单体在这里**没有直传 proto**,
// 而是取出了 affected。若将来 proto 的字段改名(如 affected →
// updated_count),HTTP 契约**不会**跟着变 —— 因为 handler 显式
// 取了那个字段名。这正是"HTTP 契约与 RPC 契约是两回事"的例子。
//
// ============================================================
// affected 的语义:被**改动**的行数,不是购物车总行数
// ============================================================
//
// 若用户已经全选,再调一次全选,affected 可能是 0(没有行发生变化)。
// 前端**不应当**用 affected == 0 判断"失败" —— 那是"无变化",
// 是正常结果。
//
// 服务端是否做"只更新有变化的行"的优化,决定 affected 是 0 还是
// 全部行数。两种都合理,故前端不该依赖它的绝对值。
func (l *SelectAllCartLogic) SelectAllCart(req *types.SelectAllCartReq) (*types.SelectAllCartResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.SelectAllCartItems(ctx, &v1_tradev1.SelectAllCartItemsReq{
		UserId:     userId,
		IsSelected: req.IsSelected,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.SelectAllCartResp{Affected: resp.GetAffected()}, nil
}
