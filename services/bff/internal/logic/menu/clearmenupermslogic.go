// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearMenuPermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearMenuPermsLogic {
	return &ClearMenuPermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClearMenuPerms 清空某菜单的权限绑定。
//
// 与 ClearRolePerms / ClearRoleMenus 同一口径:
// 与 AssignMenuPerms 传空数组功能重叠但保留两者;幂等;不预检。
//
// ============================================================
// 清空 menu_perm **不影响**任何角色的实际权限
// ============================================================
//
// 这一点值得重复(见 assignmenupermslogic.go 的详细说明):
//
//	menu_perm  决定"角色分配权限页的候选集展示什么"
//	role_perm  决定"角色实际拥有什么",判权只看它
//
// 故清空某菜单的权限绑定后:
//
//	✓ 角色分配页上,这个菜单下不再显示任何可勾选的按钮权限
//	✗ 已有角色的权限**一个都没少** —— 他们仍然拥有那些按钮
//
// 若运营的意图是"收回所有角色在这个菜单下的按钮权限",
// 那要额外去清各角色的 role_perm。**BFF 不做这个联动** ——
// 一次"清菜单绑定"的操作静默改掉几十个角色的权限,
// 是那种"点完才发现影响面"的事故。
//
// 若前端要让这个动作做到"收回权限",应当显式提示并逐角色调用
// ClearRolePerms,让影响面可见。
func (l *ClearMenuPermsLogic) ClearMenuPerms(req *types.RelIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ClearMenuPerms(ctx, &v1_userv1.ClearMenuPermsReq{
		MenuId: req.Id,
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
