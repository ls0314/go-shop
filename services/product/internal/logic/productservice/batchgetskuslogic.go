package productservicelogic

import (
	"context"
	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/model"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// batchGetSkusLimit 单次批量查询的 SKU 上限。
// 超出由调用方分批 —— 一次 RPC 拉回上万条会把 message 撑爆(默认 4MB 上限),
// 且调用方(购物车/结算)本来也只处理一页数据。
const batchGetSkusLimit = 200

type BatchGetSkusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetSkusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetSkusLogic {
	return &BatchGetSkusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// BatchGetSkus 按 SKU ID 批量查(购物车列表、结算页)。
//
// 语义:查不到的 ID 静默缺席,由调用方按"请求 - 返回"的差集判定失效项。
// 不因个别 ID 不存在而整体失败 —— 购物车里有下架商品是常态,不该让整页报错。
func (l *BatchGetSkusLogic) BatchGetSkus(in *v1_productv1.BatchGetSkusReq) (*v1_productv1.BatchGetSkusResp, error) {
	ids := in.SkuIds
	if len(ids) == 0 {
		return &v1_productv1.BatchGetSkusResp{}, nil
	}
	if len(ids) > batchGetSkusLimit {
		return &v1_productv1.BatchGetSkusResp{
			ErrorMsg: "单次批量查询的 SKU 数超过上限(200),请分批调用",
		}, nil
	}

	skuList, err := l.svcCtx.ProductRepo.GetSkuListByIds(ids)
	if err != nil {
		return nil, err
	}
	if len(skuList) == 0 {
		return &v1_productv1.BatchGetSkusResp{}, nil
	}

	// 收集 SPU ID 后一次性取回,避免 N+1
	spuIdSet := make(map[int64]struct{}, len(skuList))
	for i := range skuList {
		spuIdSet[skuList[i].SpuId] = struct{}{}
	}
	spuIds := make([]int64, 0, len(spuIdSet))
	for id := range spuIdSet {
		spuIds = append(spuIds, id)
	}
	spuList, err := l.svcCtx.ProductRepo.GetSpuListByIds(spuIds)
	if err != nil {
		return nil, err
	}
	spuMap := make(map[int64]*model.SysProductSpu, len(spuList))
	for i := range spuList {
		spuMap[spuList[i].SpuId] = &spuList[i]
	}

	items := make([]*v1_productv1.GetSkuResp, 0, len(skuList))
	for i := range skuList {
		items = append(items, converter.ToProtoGetSkuResp(&skuList[i], spuMap[skuList[i].SpuId]))
	}

	return &v1_productv1.BatchGetSkusResp{Items: items}, nil
}
