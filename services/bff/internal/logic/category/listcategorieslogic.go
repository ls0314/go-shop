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

type ListCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCategories 分页查询类目(管理端)。
//
// ============================================================
// 响应要改两处名:items → list,并补回 page / pageSize
// ============================================================
//
// proto 的 GetCategoryListResp 是 {items, total},而单体是
//
//	utils.Success(c, gin.H{
//		"list": ..., "total": ..., "page": ..., "pageSize": ...,
//	})
//
// 故 Items 必须改名成 List。**pageSize 是 camelCase** —— 与请求侧一致
// (见 convert.go),与 orders / coupons 那些域的 page_size 相反。
// 照 proto 的字段名直传会让前端拿到 undefined,**而且不报错**。
//
// page / pageSize 在 proto 响应里根本没有,是 BFF 把自己发出去的
// **兜底后**的值回写上去的。不回写前端只能拿到 0,"共 3 页"会显示成
// 1 页。
//
// ============================================================
// 列表**不过滤 status**
// ============================================================
//
// 服务端 GetCategoryList 只按 sort_order, category_id 排序,不筛状态,
// 所以 disabled 的类目也在列表里(这点与 tree 的 include_disabled
// 口径不同)。BFF 不在这一层补过滤 —— 过滤了 total 就与 list 对不上,
// 分页会错位。要按状态看,前端自己筛。
func (l *ListCategoriesLogic) ListCategories(req *types.CategoryListReq) (*types.CategoryListResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.GetCategoryList(ctx, &v1_productv1.GetCategoryListReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.CategoryListResp{
		// proto 的 items → HTTP 的 list(改名,见上)
		List:  toCategoryItems(resp.GetItems()),
		Total: resp.GetTotal(),
		// 回写的是**兜底后**的页码,不是请求里的原值(理由见 convert.go)
		Page:     page,
		PageSize: pageSize,
	}, nil
}
