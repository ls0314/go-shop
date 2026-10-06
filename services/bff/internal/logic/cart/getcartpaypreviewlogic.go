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

type GetCartPayPreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCartPayPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartPayPreviewLogic {
	return &GetCartPayPreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCartPayPreview 结算预览:仅统计**已选中**的行。
//
// ============================================================
// proto 只给一半字段,另一半由 BFF 算
// ============================================================
//
// proto 的 CartPayPreview 只有三个字段:
//
//	repeated CartItem selected_items = 1;
//	int64 total_quantity = 2;
//	double total_amount = 3;
//
// 而 HTTP 契约(单体 response.CartItemPayResp)有六个。已核实单体的
// 转换函数 tradeclient.toModelPayPreview(与本节实现一一对应):
//
//	available   := items 里 IsAvailable 的那些
//	unavailable := items 里 !IsAvailable 的那些
//
//	Items            = available        ← **只含可购买的!**
//	TotalCount       = len(items)       ← 总数(可 + 不可)
//	TotalQuantity    = proto 直传
//	TotalAmount      = proto 直传
//	HasUnavailable   = len(unavailable) > 0
//	UnavailableItems = unavailable
//
// ============================================================
// **items 不含不可购买的行** —— 这是最容易搞错的一处
// ============================================================
//
// 直觉上 items 应当是"全部选中项",不可购买的既在里面、也在
// unavailable_items 里(重复出现)。但单体不是这样:
//
//	它把不可购买的**从 items 里移出去了**,只在 unavailable_items 里出现。
//
// 后果差别很具体:前端渲染"待结算商品清单"时遍历 items ——
// 若 BFF 把不可购买的也放进 items,用户会在待结算清单里看到
// 一个买不了的商品,而它同时又出现在下面的"无法购买"区块,
// **同一行显示两次**。
//
// 单体的注释也说明了这个切分为什么在调用方做:
//
//	"这个切分在**调用方**做而不是服务端:服务端只描述'每行是否
//	 可购买',怎么展示是前端的决定"
//
// ============================================================
// total_count 与 total_quantity 的含义不同
// ============================================================
//
//	total_count    = **种数**(可 + 不可,即 proto 给的全部选中行)
//	total_quantity = **件数**(proto 注释:"选中商品总件数")
//
// 注意 total_count 数的是**全部选中行**,而 items 只有可购买的那些
// —— 所以 `total_count != len(items)` 在存在不可购买项时会成立。
// 这是刻意的:它表达"你选中了 N 种",而 items 表达"其中这些能买"。
//
// 另有一个容易混的:用户端购物车的 GetCartItemTotal 返回的 count
// 也是"种数",但它数的是**整个购物车**(不限选中)。
func (l *GetCartPayPreviewLogic) GetCartPayPreview(req *types.Empty) (*types.CartPayPreviewResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.GetCartPayPreview(ctx, &v1_tradev1.GetCartPayPreviewReq{
		UserId: userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	preview := resp.GetPreview()
	// selected_items → 全部选中项(可 + 不可)
	selected := toCartItems(preview.GetSelectedItems())

	// 按可购买性切分 —— 与单体 toModelPayPreview 的算法一致。
	//
	// 两个都用 make(..., 0, cap) 而不是 var x []T:空时是 [] 而不是
	// null,前端 map/length 不会抛错。
	available := make([]types.CartItem, 0, len(selected))
	unavailable := make([]types.CartItem, 0)
	for _, it := range selected {
		if it.IsAvailable {
			available = append(available, it)
		} else {
			unavailable = append(unavailable, it)
		}
	}

	return &types.CartPayPreviewResp{
		// **只含可购买的**(见上方说明)
		Items: available,
		// 全部选中行的**种数**(与 len(Items) 可能不等)
		TotalCount:    int64(len(selected)),
		TotalQuantity: preview.GetTotalQuantity(),
		TotalAmount:   preview.GetTotalAmount(),
		// 用 len 判断而不是单独维护一个 bool,保证两者永远一致
		HasUnavailable:   len(unavailable) > 0,
		UnavailableItems: unavailable,
	}, nil
}
