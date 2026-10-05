// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearRolePermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearRolePermsLogic {
	return &ClearRolePermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClearRolePerms 清空某角色的权限绑定。
//
// 与 ClearRoleMenus 同一口径(见那边的详细说明):
// 与 AssignRolePerms 传空数组功能重叠但保留两者;幂等;不预检。
//
// ============================================================
// 清空权限**不影响**该角色已绑定的用户与菜单
// ============================================================
//
// 这三组绑定(角色-权限 / 角色-菜单 / 用户-角色)是相互独立的表:
//
//	清空角色权限 → 该角色下的用户变成"能进页面但没有任何按钮权限"
//	清空角色菜单 → 该角色下的用户侧边栏变空
//	清空角色本身 → 连带清掉它的三组绑定
//
// 前两者**不会**级联。这是刻意的 —— 管理员可能正在重建权限
// (先清后配),中间态不该把用户的角色也摘掉。
//
// 服务端的行为若与上述不符(比如清权限时顺带删了菜单绑定),
// 那是一个 bug,而不是 BFF 要在这里补偿的东西。
func (l *ClearRolePermsLogic) ClearRolePerms(req *types.RelIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ClearRolePerms(ctx, &v1_userv1.ClearRolePermsReq{
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.Empty{}, nil
}
