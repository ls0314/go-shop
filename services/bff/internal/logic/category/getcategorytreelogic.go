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

type GetCategoryTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryTreeLogic {
	return &GetCategoryTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCategoryTree 类目树。
//
// ============================================================
// 查询参数是 snake_case:level / include_disabled
// ============================================================
//
// 与同域的列表接口(pageSize,camelCase)**相反** —— 一个域里两套命名。
// 这不是笔误:前端 CategoryTreeParams 就是 snake_case,单体那边也是
// c.DefaultQuery("level","0") / ("include_disabled","false")。不要统一。
//
// level 的口径在服务端:0(或不传)= 全部层级,1~4 = 只取该层,而
// **>4 不过滤**(repository/category.go 的条件是 level > 0 && level <= 4,
// 传 5 会拿到全部)。BFF 不在这里加校验 —— 加了就会出现"BFF 认为非法、
// 服务端照做"的第二套规则。
//
// ⚠️ 按 level 过滤是**平铺**过滤,不是"取前 N 层":level=3 时父类目不在
// 结果集里,服务端会把这类悬空节点提为根(见 makeCategoryTree),前端
// 拿到的是一片森林而不是完整三层树。要完整树就别传 level。
//
// ============================================================
// 没有 parent_id,children 递归
// ============================================================
//
// 树里父节点由层级隐含,单体在 GetTreeCategoryResp 上标了 json:"-",
// proto 的 CategoryTreeNode 同样没有 parent_id —— 两者一致,故
// types.CategoryTreeItem 也没有这个键。转换细节见 convert.go。
func (l *GetCategoryTreeLogic) GetCategoryTree(req *types.CategoryTreeReq) ([]types.CategoryTreeItem, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.GetCategoryTree(ctx, &v1_productv1.GetCategoryTreeReq{
		Level:           req.Level,
		IncludeDisabled: req.IncludeDisabled,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// 一条类目都没有时返回 [] 而不是 null(见 convert.go)
	return toCategoryTreeItems(resp.GetItems()), nil
}
