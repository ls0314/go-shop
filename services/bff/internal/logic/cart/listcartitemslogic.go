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

type ListCartItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCartItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCartItemsLogic {
	return &ListCartItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCartItems 取当前用户的购物车列表。
//
// ============================================================
// 响应是**裸数组**,不是 {list: [...]}
// ============================================================
//
// 单体: utils.Success(c, resp) —— 而 resp 是
// []response.CartItemListResp。故 data 直接就是数组:
//
//	{ "code":200, "message":"Success", "data": [ {...}, {...} ] }
//
// 故 .api 里这条是 `returns ([]CartItem)`。
//
// **注意与 GetCartPayPreview 的差别**:那条是
// utils.Success(c, resp),而 resp 是 **struct**(response.CartItemPayResp),
// 所以它有 items 键。两者都是"列表",但一个是裸数组、一个是包装 ——
// 这正是"不能跨接口类推形状"的例子。
//
// ============================================================
// 不分页,这是刻意的
// ============================================================
//
// proto 的注释写明了理由:
//
//	"购物车不分页:一个用户的购物车行数天然有限,分页会让
//	 '全选/合计'这类操作要跨页统计,反而更容易算错。"
//
// 故这条没有 page/page_size,也没有 total —— 前端拿到整个数组,
// 全选与合计都在本地算。
//
// ============================================================
// 商品侧字段是**实时回填**的
// ============================================================
//
// proto 的注释:"商品侧的展示字段(名称/图/规格/价格/库存)在读写时经
// product-service 的 BatchGetSkus 回填,购物车不存商品快照。"
//
// 故每一行的 price / stock / is_available 都是**此刻**的值,
// 而不是加购那一刻的。这是购物车的正确语义(价格变了要显示新价),
// 但也意味着这个接口会打 product-service 的批接口 —— 它的耗时
// 受下游影响,2 秒的调用超时对大批量购物车可能偏紧。
func (l *ListCartItemsLogic) ListCartItems(req *types.Empty) ([]types.CartItem, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.ListCartItems(ctx, &v1_tradev1.ListCartItemsReq{
		UserId: userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// 空购物车返回**空数组**而不是错误 —— 这是正常的业务状态
	// (与"部门树在用户无所属部门时报错"不同)。
	//
	// toCartItems 用 make(..., 0, len),故空时是 [] 而不是 null。
	return toCartItems(resp.GetItems()), nil
}
