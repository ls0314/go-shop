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

type WithdrawProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawProductLogic {
	return &WithdrawProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WithdrawProduct 下架(状态机只接受 published → withdrawn)。
//
// ============================================================
// 下架与上架的校验强度**不对称**,这是刻意的
// ============================================================
//
//	上架:校验类目 + SKU + 库存 + 价格(要把商品放进橱窗,必须能卖)
//	下架:只校验当前状态是 published
//
// 下架是一次"止损"操作 —— 商品可能因为库存清零、类目被停用、价格配错
// 而必须立刻撤下。若下架也要求"类目可用、有 active SKU",那些**正需要
// 下架的商品反而下不了架**,只能干看着。故服务端对下架只做状态判定,
// BFF 更不该在这里加条件。
//
// ============================================================
// 下架不等于删除
// ============================================================
//
// 下架只是 spu_status = withdrawn:
//
//	管理端列表仍然看得到(能看到才能再上架)
//	用户端列表/详情查不到(只返回 published)
//	SKU / 库存 / 历史订单全部保留
//
// 要彻底不可见用 DeleteProduct(软删)。
//
// 另一处容易混的:下架之后**不能直接改规格模板**。服务端的局部更新只
// 允许 withdrawn → draft 这一种状态转换,要把商品重新编辑上架,
// 路径是:下架 → 改回 draft → 修改 → 再上架。
func (l *WithdrawProductLogic) WithdrawProduct(req *types.ProductIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.WithdrawProduct(ctx, &v1_productv1.WithdrawProductReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 当前不是 published(含重复下架)→ 400。
		// 重复下架报错而不是静默成功,理由同上架(FIX-3)。
		return nil, err
	}

	return &types.Empty{}, nil
}
