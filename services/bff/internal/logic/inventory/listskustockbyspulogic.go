// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package inventory

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSkuStockBySpuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSkuStockBySpuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSkuStockBySpuLogic {
	return &ListSkuStockBySpuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListSkuStockBySpu 某个 SPU 下全部 SKU 的库存 + 三个聚合值。
//
// 响应形状 {list, total_stock, total_lock, total_sold} 里有两处必须留意:
//
//	items → list   改名的老规矩(repeated items 是 proto 的说法,
//	               HTTP 契约里一律叫 list)。
//
//	total_stock    **与列表里每一行的 total_stock 不是一回事**:
//	               顶层的三个聚合值全部由 product-service 算好
//	               (Σ可用 / Σ锁定 / Σ销量),BFF 只透传;
//	               而每一行的 total_stock 是"该 SKU 的可用+锁定",
//	               由 BFF 推导。故 Σ 每行的 total_stock ≠ 顶层值。
//	               完整推导见 convert.go 顶部,不要在这里"顺手修正"。
//
// SPU 不存在时服务端回 error_msg(整体失败,而不是给一个空列表)——
// 那样前端会看到"这个 SPU 一个 SKU 都没有",更难排查。
func (l *ListSkuStockBySpuLogic) ListSkuStockBySpu(req *types.SpuStockReq) (*types.SpuStockResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.InventoryRPC.GetSkuStockList(ctx, &v1_productv1.GetSkuStockListReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// make(..., 0, len) 而不是 var x []T:该 SPU 下没有 SKU 时下发 []
	// 而不是 null,前端 map/length 不会抛错。
	items := resp.GetItems()
	list := make([]types.SkuInventory, 0, len(items))
	for _, it := range items {
		list = append(list, toSkuInventory(it))
	}

	return &types.SpuStockResp{
		List:       list,
		TotalStock: resp.GetTotalStock(),
		TotalLock:  resp.GetTotalLock(),
		TotalSold:  resp.GetTotalSold(),
	}, nil
}
