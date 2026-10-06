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

type GetCartCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCartCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartCountLogic {
	return &GetCartCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCartCount 取购物车总数量(导航栏角标用)。
//
// ============================================================
// 响应是 {count: N},而 RPC 里那个字段叫 total
// ============================================================
//
// proto 的 GetCartItemCountResp 是 {total, error_msg},而单体
// handler 手工改名成 count:
//
//	utils.Success(c, gin.H{"count": total})
//
// 故 .api 里是 types.CartCountResp{Count},**不是** {Total}。
// 照 proto 写成 total 会让导航栏角标显示不出来(前端解的是
// res.data.count)。
//
// ============================================================
// 这个接口的语义:商品**种类数**,不是件数
// ============================================================
//
// proto 注释与单体 handler 的注释都写明了:
//
//	proto:  "购物车数量(不是件数)"
//	handler: "获取当前用户购物车中的总商品种类数(用于导航栏角标展示)"
//
// 即 3 种商品各买 5 件,这里返回 **3** 而不是 15。
//
// 这与 GetCartPayPreview 的 total_quantity 不同 —— 那个是**件数**
// (5+5+5=15)。两者名字像(都是"数量")但含义不同,读代码时
// 极易混。前端若拿这个值显示"共 N 件"就错了。
//
// ============================================================
// 为什么不复用 ListCartItems 的返回长度
// ============================================================
//
// 理论上 count == len(list),但:
//
//	① 这条接口是给角标用的,可能在每个页面都调 —— 让它去打
//	   product-service 的 BatchGetSkus(列表接口会)是浪费;
//	② 服务端可以直接 count(*) 而不回填商品字段,便宜得多。
//
// 故它是**独立的轻量接口**,不是列表的派生。
func (l *GetCartCountLogic) GetCartCount(req *types.Empty) (*types.CartCountResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CartRPC.GetCartItemCount(ctx, &v1_tradev1.GetCartItemCountReq{
		UserId: userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// proto 的 total → HTTP 的 count(见上方说明)
	return &types.CartCountResp{Count: resp.GetTotal()}, nil
}
