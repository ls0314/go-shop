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

type PublishProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishProductLogic {
	return &PublishProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublishProduct 上架(状态机只接受 draft / withdrawn → published)。
//
// ============================================================
// 上架有一串前置校验,全在服务端
// ============================================================
//
//	类目存在、is_leaf、status = active   —— 否则"挂在不存在的类目下"
//	至少一个 sku_status = active 的 SKU
//	该 SKU 库存 > 0
//	该 SKU 价格 > 0
//
// BFF **不做任何预检**。它要先查商品、查类目、查 SKU 才能判,那是三次
// 跨服务读 + 一次写,而且查完到写之间别人可能改了库存(TOCTOU)。
// 服务端把校验与写在**同一个事务**里做,那才是唯一正确的判定方。
//
// ============================================================
// 重复上架会**报错**(这是与单体行为不同的一处修正)
// ============================================================
//
// 服务端更新时带 `spu_status IN ('draft','withdrawn')` 条件,命不中即
// "状态转换不合法";单体那边同样写了这个判断但**整段被吞掉**(FIX-3),
// 于是对已上架商品重复上架会返回"成功"而状态没变 —— 前端据此提示
// "上架成功",运营却看不到任何变化。
//
// 故这里必须把 error_msg 照实回 400,不要为了让前端"幂等"而吞掉它。
//
// ============================================================
// 副作用:上架后会同步 ES
// ============================================================
//
// 服务端在事务提交后清详情缓存并同步搜索索引(事务回滚则不同步)。
// 故"上架成功"之后商品应当能在用户端列表搜到 —— 若搜不到,
// 那是 ES 同步的问题,不是 BFF 的(端到端验证时按这个顺序排查)。
func (l *PublishProductLogic) PublishProduct(req *types.ProductIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.PublishProduct(ctx, &v1_productv1.PublishProductReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 状态转换不合法 / 类目不可用 / 没有可用 SKU 或库存为 0 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
