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

type AssignMenuPermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignMenuPermsLogic {
	return &AssignMenuPermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssignMenuPerms 全量替换菜单的权限绑定。
//
// **全量替换,不是追加**(与其它四组同一语义):传的必须是完整集合,
// 原本 [1,2] 想加 3 要传 [1,2,3]。
//
// 空数组合法(表示清空),不拦 —— 前端"全部取消勾选后保存"会用到。
//
// ============================================================
// 菜单-权限 与 角色-权限 是两条独立的链路
// ============================================================
//
// 这两组绑定常被混为一谈,但它们的用途完全不同:
//
//	menu_perm   定义"这个菜单下**有哪些**权限点可用"
//	            → 角色分配权限页据此渲染:勾了这个菜单,才显示
//	              它下面的按钮权限供勾选
//	role_perm   定义"这个角色**实际拥有**哪些权限点"
//	            → 判权时只看它
//
// 所以 menu_perm 变了**不会**自动改变任何角色的权限 ——
// 它只是改变了"可选项的展示范围"。
//
// 反过来说:若某个权限点从 menu_perm 里被移除,而已有角色的
// role_perm 里还绑着它,那个角色**仍然拥有**它(判权看 role_perm)。
// 这可能是也可能不是想要的 —— 但那是运营要去清理的数据,
// BFF 不做隐式联动(隐式联动会让"改菜单"这个动作产生
// 难以预期的权限变更)。
//
// 响应回显请求体:见 role 包 assignrolepermslogic.go 的说明
// (单体是 utils.Success(c, req),BFF 回 {} —— 已知可见差异)。
func (l *AssignMenuPermsLogic) AssignMenuPerms(req *types.AssignMenuPermsReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.AssignMenuPerms(ctx, &v1_userv1.AssignMenuPermsReq{
		MenuId:  req.MenuId,
		PermIds: req.PermIds,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 菜单不存在 / 权限 ID 不存在 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
