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

type GetCategoryChildrenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryChildrenLogic {
	return &GetCategoryChildrenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCategoryChildren 查某类目的直接子类目。
//
// ============================================================
// include_disabled 在 BFF 这一层被丢掉 —— 它是一个**无效参数**
// ============================================================
//
// proto 的 GetCategoryChildrenReq 只有 parent_id,**没有** include_disabled
// (单体那侧的 GetCategoryChildrenList 也只传 id)。故 req.IncludeDisabled
// 无处可传:前端传 true / false 都不改变结果。
//
// 更要紧的是服务端的行为:它用 GetCategoryByParentId,SQL 是
// `WHERE parent_id = ?`,**不筛 status** —— 也就是**永远包含已停用的
// 子类目**,与 include_disabled=false 的直觉正好相反。前端若用这个参数
// 来控制"停用的子类目不要显示",会看到它们照旧返回。
//
// 真正的修法是给 proto / 服务端补字段并改 SQL —— BFF 单方面过滤会造成
// "children 少了几个、但 tree 里还有"的不一致,比现在更难查。故这里
// 如实转发,不做补偿。
//
// ============================================================
// "父类目不存在"与"没有子类目"是两回事
// ============================================================
//
// 前者服务端返回 CategoryNotExist → 400;后者是 200 + [](空数组,不是
// null)。前端据此区分"这个 ID 是错的"和"这个类目底下是空的"。
//
// 父类目已停用也能查到子类目(服务端只判存在性),故这里不做额外前置
// 校验。
func (l *GetCategoryChildrenLogic) GetCategoryChildren(req *types.CategoryChildrenReq) ([]types.CategoryListItem, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.GetCategoryChildren(ctx, &v1_productv1.GetCategoryChildrenReq{
		// 路径参数 :id 是**父类目** id
		ParentId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 父类目不存在 → 400
		return nil, err
	}

	return toCategoryItems(resp.GetItems()), nil
}
