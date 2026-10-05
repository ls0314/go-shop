// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRolesLogic {
	return &ListUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListUserRoles 取某用户已绑定的角色(角色分配页回显已选)。
//
// ============================================================
// 路径参数是 :id 但语义是 user_id
// ============================================================
//
// 单体路由是 GET /api/v1/admin/user/:id/role,.api 里沿用 :id
// (types.RelIdReq)。而 RPC 的入参叫 user_id ——
// 因为"谁的绑定"里的"谁"是用户。
//
// 这个映射不显然,故写出来:req.Id → UserId。
//
// ============================================================
// 响应形状:items / total(不是 list)
// ============================================================
//
// 与 ListUsers 不同 —— 这条是 proto 直传(ListUserRolesResp),
// 字段名 items/total。
//
// **同一个路径前缀下的两个接口,键名不一样**,这是既有契约:
//
//	GET /admin/user         → list / total / page / pageSize
//	GET /admin/user/:id/role → items / total
//
// 前者是单体 handler 内联拼的 gin.H,后者是 proto 直传。
// 迁移时**不能统一**,前端按各自的键解。
func (l *ListUserRolesLogic) ListUserRoles(req *types.RelIdReq) (*types.ListUserRolesResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListUserRoles(ctx, &v1_userv1.ListUserRolesReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListUserRolesResp{
		List:  converter.RoleItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
