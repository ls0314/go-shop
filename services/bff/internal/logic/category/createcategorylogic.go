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

type CreateCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateCategory 创建类目。
//
// ============================================================
// level / path / is_leaf 由服务端推导,BFF 一个都不碰
// ============================================================
//
// 请求里没有 category_level / category_path / is_leaf / status:前三个由
// 服务端按父类目算(父类目查不到 → CategoryParentNotExist,父层级 ≥4 →
// CategoryLevelDeep,创建后父类目的 is_leaf 会被置 false),status 缺省
// 为 active。BFF 若"顺手"补一个 level,只会与权威算出来的值打架。
//
// ============================================================
// 与单体的两处差别(都是 .api 定稿时定下的)
// ============================================================
//
//	单体把 body 直接绑成 model.SysCategory,所以**建的时候就能带
//	status**,一次请求建出 disabled 类目;BFF 的 CreateCategoryReq 没有
//	status 字段,建出来一律是服务端默认的 active,要停用得再走一次
//	PUT(updates 里带 "status":"disabled")。
//
//	is_visible 在创建路径上**没有**更新路径那个取舍问题:创建是整体传
//	对象,不存在"传了才改",不传即 false —— 建出来的类目是隐藏的,
//	与单体的绑定行为一致。
//
// 业务失败(同父下重名 CategoryUkExist / 父类目不存在 / 层级过深 /
// 父类目已停用)由服务端以 error_msg 返回 → 400,原因原样透传给前端。
func (l *CreateCategoryLogic) CreateCategory(req *types.CreateCategoryReq) (*types.CreateCategoryResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.CreateCategory(ctx, &v1_productv1.CreateCategoryReq{
		Category: &v1_productv1.Category{
			ParentId:     req.ParentId,
			CategoryName: req.CategoryName,
			SortOrder:    req.SortOrder,
			IconUrl:      req.IconUrl,
			IsVisible:    req.IsVisible,
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// proto 的 CreateCategoryResp 只回 category_id 与 category_path
	// (加 error_msg),**不回完整对象** —— 想拿全字段得再 GET 一次。
	// 前端创建后通常只需要跳转到 /category/:id,故契约如此即可。
	return &types.CreateCategoryResp{
		CategoryId:   resp.GetCategoryId(),
		CategoryPath: resp.GetCategoryPath(),
	}, nil
}
