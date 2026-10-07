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

type GetCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCategory 类目详情。
//
// 十个字段与 proto 的 Category **一一对应**(category_id / parent_id /
// category_name / category_level / category_path / sort_order / icon_url /
// is_leaf / is_visible / status),单体 GetCategoryResp 也是这十个 ——
// 类目域没有 cart 那种"改名 + 改类型 + 隐藏"的转换,只是搬运。
//
// 查不到时服务端回 error_msg(CategoryNotExist)→ 400,不是 200 + null:
// 编辑页据此区分"这条类目被删了"与"网络/服务故障"。
//
// 类目是全局数据,不按 user_id 过滤,故这里不取身份 —— 能进到这个
// logic 说明请求已过 Auth 中间件(路由已挂),鉴权在网关那层做完。
func (l *GetCategoryLogic) GetCategory(req *types.CategoryIdReq) (*types.Category, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.GetCategory(ctx, &v1_productv1.GetCategoryReq{
		CategoryId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 类目不存在 → 400
		return nil, err
	}

	return toCategory(resp.GetCategory()), nil
}
