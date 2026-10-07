// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteProduct 删除商品(**软删 SPU 并级联软删其 SKU**)。
//
// ============================================================
// 是软删,不是物理删
// ============================================================
//
// 服务端在同一个事务里:校验 SPU 存在且未删 → 批量软删 SKU →
// 软删 SPU(is_deleted = true),提交后再清详情缓存、从 ES 移除。
//
// 这解释了三件事:
//
//	历史订单不受影响 —— 订单详情里的商品名/规格是**下单时的快照**,
//	                      本来就不回查商品表;而且行还在,只是标记位变了
//	外键不会挡住删除 —— 建表时 fk_sku_spu_sku / fk_spu_category_spu 都是
//	                      ON DELETE RESTRICT(不是 CASCADE),但软删只改
//	                      字段,不触发外键动作。若哪天真做物理删,
//	                      这两个 RESTRICT 会立刻挡住(即"类目下有商品不许删类目")
//	同一商品删两次会报错 —— 第二次拿到的 SPU 已是 is_deleted = true,
//	                      服务端按"商品不存在"回 error_msg → 400
//
// ============================================================
// 删之前没有"是否有关联订单"的检查 —— 已知缺口
// ============================================================
//
// 服务端那里留着一行待办注释:订单关联检查要等订单域拆出来之后才能做
// (订单表在 trade 侧,product 跨库取不到)。所以现在**删掉一个卖出过的
// 商品是允许的**。
//
// 后果有限(软删 + 订单快照,用户的历史订单照常显示),但"销量/评价类
// 报表"会失去商品维度。这是跨域问题,BFF 这一层解决不了,不要在 logic
// 里加"先查订单再删"的伪预检 —— 它要跨服务调用,而且是 TOCTOU。
func (l *DeleteProductLogic) DeleteProduct(req *types.ProductIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.DeleteProduct(ctx, &v1_productv1.DeleteProductReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 已删除 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
