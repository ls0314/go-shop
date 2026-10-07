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

type ListProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListProducts 管理端商品分页列表。
//
// ============================================================
// 同一条接口的两侧分页命名**不一致** —— 照实实现,不要"统一"
// ============================================================
//
//	请求  page_size   (snake_case,.api 的 form tag;前端 SpuQueryReq 也是它)
//	响应  pageSize    (camelCase,.api 的 json tag;前端 GetSpuListResp 也是它)
//
// 这是单体遗留:请求走结构体的 form tag,响应是手写的 gin.H。
// 把它"顺手统一"成一种写法,前端必然有一侧拿到 undefined ——
// 请求侧会变成"永远只返回默认 10 条",响应侧会是"分页器算不出总页数",
// 而且两种都**不报错**。
//
// ============================================================
// items → list
// ============================================================
//
// proto 的 repeated 字段叫 items,HTTP 契约叫 list。proto 直传会让前端
// 拿到 undefined(list.length 直接抛错)。
//
// ============================================================
// category_id:零值即不过滤
// ============================================================
//
// proto 是 Int64Value(可选,proto 注释:"为空表示不过滤"),HTTP 是普通
// int64 —— "没传"与"传了 0"在 HTTP 侧已经不可区分,故 0 一律当不过滤
// (见 optionalInt64 的说明与代价)。
//
// ============================================================
// 分页兜底与封顶在**服务端**,响应里的 page/pageSize 是归一化后的值
// ============================================================
//
// product-service 自己把 page<=0 → 1、pageSize<=0 → 10、pageSize 封顶
// (管理端 100、用户端 50),并把**归一化后**的值回填进响应;单体也是直接
// 用响应里的这两个值回给前端,故 BFF 照做。
//
// 这里**刻意不再兜一次**:两处各兜一次,将来改默认值时只会改一处,
// 于是"BFF 以为的默认值"与"服务端实际用的默认值"分叉,症状是分页器
// 显示的条数与实际返回条数对不上 —— 这类不一致很难查。
func (l *ListProductsLogic) ListProducts(req *types.AdminProductListReq) (*types.AdminProductListResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.GetProductList(ctx, &v1_productv1.GetProductListReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		SpuName:  req.SpuName,
		// 零值 → nil(不过滤)
		CategoryId: optionalInt64(req.CategoryId),
		// spu_status 是管理端**独有**的筛选:管理端要看草稿/已下架,
		// 用户端只可能看到已上架(服务端硬编码),故那边的请求里没有这个字段。
		SpuStatus: req.SpuStatus,
		Brand:     req.Brand,
		Sort:      req.Sort,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品引用了取不到名称的类目(脏引用/类目被删)→ 400
		return nil, err
	}

	return &types.AdminProductListResp{
		// items → list,并顺带把 spec_* 之外的类型差异处理掉(见 convert.go)
		List:  toAdminSpuListItems(resp.GetItems()),
		Total: resp.GetTotal(),
		// **camelCase 的 pageSize**,见上方说明
		Page:     int(resp.GetPage()),
		PageSize: int(resp.GetPageSize()),
	}, nil
}
