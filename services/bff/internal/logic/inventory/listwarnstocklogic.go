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

type ListWarnStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListWarnStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWarnStockLogic {
	return &ListWarnStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListWarnStock 可用库存低于阈值的 SKU 列表(管理端预警),按库存升序。
//
// 响应是**裸数组**:单体 handler 是 utils.Success(c, warnStockList),
// data 直接就是 [ {...}, {...} ],没有 {list, total} 外壳。故 .api 里
// 这条 returns ([]WarnStockItem),生成 (resp []types.WarnStockItem, ...)。
// proto 侧的 items 在这里被整个拆掉 —— 与 GetSkuStock 拆 sku 那层
// 包装是同一类处理。
//
// 空结果必须是 [] 而不是 null:前端直接对 data 做 .map()/length,
// nil 切片会序列化成 null 并把它打挂(见下面的 make)。
//
// ============================================================
// 默认值不在 BFF
// ============================================================
//
// threshold <= 0 时服务端取 10,spu_status 空时服务端取 published
// (product-service 与单体同口径)。BFF **原样透传**、不在这里填默认值:
// 填了就有两份默认值,以后调整阈值口径时必然只改一处。
//
// 另注:预警条目里没有 spec_values,也没有 total_stock —— 它只列
// sku_id/名称/库存三个数与状态,故本 logic 不需要 parseSpecValues。
func (l *ListWarnStockLogic) ListWarnStock(req *types.WarnStockReq) (resp []types.WarnStockItem, err error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	respRPC, grpcErr := l.svcCtx.InventoryRPC.GetWarnStockList(ctx, &v1_productv1.GetWarnStockListReq{
		// proto 那边是 int32,HTTP 契约里是 int64(表单里的整数都是
		// 64 位解析)—— 预警阈值不会超出 int32,直接窄化。
		Threshold: int32(req.Threshold),
		SpuStatus: req.SpuStatus,
	})

	// GetWarnStockListResp 有 error_msg(第 2 字段)。
	kind, err := rpc.Classify(respRPC.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	items := respRPC.GetItems()
	list := make([]types.WarnStockItem, 0, len(items))
	for _, it := range items {
		list = append(list, types.WarnStockItem{
			SkuId:     it.GetSkuId(),
			SpuName:   it.GetSpuName(),
			SkuName:   it.GetSkuName(),
			Stock:     it.GetStock(),
			LockStock: it.GetLockStock(),
			SoldCount: it.GetSoldCount(),
			SkuStatus: it.GetSkuStatus(),
		})
	}

	return list, nil
}
