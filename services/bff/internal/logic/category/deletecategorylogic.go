// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package category

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteCategory 删除类目。
//
// ============================================================
// 拒绝条件是"有子类目 / 已被商品关联",判定全在服务端
// ============================================================
//
// 服务端在事务里查子类目(CategoryHasChildren)与商品关联(CategoryHasRel),
// 命中就把 error_msg 回给 BFF → 400,前端拿到的原因很具体。
//
// BFF 不做"先查子类目再删"的预检:那是多一次 RPC、且两次之间状态可能
// 变化(TOCTOU),还会与真正的权威漂移 —— 与 role 域 DeleteRole 同一
// 理由(那边是"系统角色不可删")。
//
// 顺带一提:删掉一个叶子类目后服务端会把父类目重新判为 is_leaf
// (同父下已无其它子类目时),这一步也在服务端完成。
//
// ============================================================
// 响应:单体是 null,这里是 {}
// ============================================================
//
// 单体: utils.Success(c, nil) → "data": null
// BFF:  returns (Empty) → &types.Empty{} → "data": {}
//
// 这是全项目 returns (Empty) 路由的共同差异(多处已有标注)。若前端
// 判 `res.data === null` 会走错分支 —— 端到端验证时确认。
//
// 幂等性由服务端决定:BFF 只透传它的 error_msg(类目不存在 → 400)。
// 前端"删除"按钮通常不关心"已删过"与"删成功"的区别。
func (l *DeleteCategoryLogic) DeleteCategory(req *types.CategoryIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.DeleteCategory(ctx, &v1_productv1.DeleteCategoryReq{
		CategoryId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 类目不存在 / 有子类目 / 已关联商品 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
