// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package inventory

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdjustStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdjustStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustStockLogic {
	return &AdjustStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdjustStock 手动调整库存(管理端)。
//
// ============================================================
// 操作人取自 JWT,不在请求体里
// ============================================================
//
// types.AdjustStockReq 只有 sku_id / change_qty / remark,而 proto 的
// AdjustStockReq 多一个 user_id —— 它落到库存流水的 create_by 上。
//
// 为什么不放进请求体:那是**审计字段**,由客户端自报就等于让调用方
// 决定"这笔记在谁头上"。单体也是从 JWT 上下文取(handler 里的
// GetUserInfoByContext)。故这里用 middleware.UserID。
//
// 取不到身份只有一种可能:该路由没挂 Auth 中间件(配置错误),回
// errNoIdentity → 500,让运维看见,而不是静默用 user_id=0 记一笔
// "系统调整"。
//
// ============================================================
// remark 不在这里校验
// ============================================================
//
// 服务端对空 remark 是**拒绝**(model.ErrRemarkEmpty,审计要求),
// 经 error_msg 回来。BFF 不重复校验:两处都判,以后放宽规则时必然
// 只改一处,而 BFF 这一份会变成"用户改不了、后端也不报错"的隐形限制。
//
// 注意状态码:单体这里把 error_msg 一律回 500(utils.Error(c,500,...)),
// 而 BFF 按统一口径把 error_msg 归为业务失败 → 400(response.Failure)。
// 这是整个 BFF 的既有取舍,不在本域单独开例外。
//
// ============================================================
// 响应只有前后库存
// ============================================================
//
// 幂等命中(同一调整重复提交)时 before == after —— 服务端的语义是
// "已经调过了",不报错。BFF 原样透传,不额外提示。
func (l *AdjustStockLogic) AdjustStock(req *types.AdjustStockReq) (*types.AdjustStockResp, error) {
	userID, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.InventoryRPC.AdjustStock(ctx, &v1_productv1.AdjustStockReq{
		SkuId:     req.SkuId,
		ChangeQty: req.ChangeQty,
		Remark:    req.Remark,
		// 落库存流水的 create_by
		UserId: userID,
	})

	// AdjustStockResp 有 error_msg(第 3 字段):remark 为空 / SKU 不存在 /
	// 调完为负都是业务失败。
	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.AdjustStockResp{
		BeforeStock: resp.GetBeforeStock(),
		AfterStock:  resp.GetAfterStock(),
	}, nil
}
