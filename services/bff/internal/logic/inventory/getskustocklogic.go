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

type GetSkuStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSkuStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkuStockLogic {
	return &GetSkuStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSkuStock 单个 SKU 的库存详情(管理端库存看板)。
//
// 响应是**裸对象**而不是 {list: [...]}:单体 handler 是
// utils.Success(c, skuStock),故 .api 里这条 returns (SkuStockResp)。
// proto 那边的形状是 GetSkuStockResp{sku: SkuStock} —— 一层包装,
// 这里要把它拆掉。
//
// total_stock 由 BFF 推导(proto 的 SkuStock 没有这个字段),
// 定义与依据见 convert.go 顶部的说明。
//
// 不判权:与其它管理端库存接口一样只做认证(中间件挂 Auth),
// 权限点由单体的 PermissionMiddleware 负责,BFF 这一版不接管。
func (l *GetSkuStockLogic) GetSkuStock(req *types.SkuStockReq) (*types.SkuStockResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.InventoryRPC.GetSkuStock(ctx, &v1_productv1.GetSkuStockReq{
		SkuId: req.Id,
	})

	// GetSkuStockResp 有 error_msg(第 2 字段):SKU 不存在、SPU 已删
	// 都是服务端用 error_msg 表达的"业务失败"。
	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// 服务端既没报 grpc 错也没给 error_msg,却回了个空 sku —— 正常情况下
	// 不会发生(见 getskustocklogic 的返回路径)。此时返回 nil,序列化成
	// data:null,与单体 toModelSkuInventory(nil) → utils.Success(c, nil)
	// 一致;不编造零值对象(见 convert.go 里 toSkuStockResp 的说明)。
	return toSkuStockResp(resp.GetSku()), nil
}
