// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMenuPermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenuPermsLogic {
	return &ListMenuPermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListMenuPerms 取某菜单已绑定的权限点。
//
// 路径参数 :id 的语义是 **menu_id**(路由是 /admin/menu/:id/perm)。
//
// 五组关联绑定的 list 里,只有这一组和 ListRolePerms 的元素是
// **权限**(另外三组分别是 Role / Menu / Dept)。
//
// ============================================================
// 这个列表的用途:渲染"菜单下的按钮权限"候选集
// ============================================================
//
// 角色分配权限页的交互通常是:
//
//	① 左树:菜单(GET /admin/menu/tree)
//	② 选中某菜单 → 中间列:该菜单下的权限点(**本接口**)
//	③ 右侧:该角色已勾选的权限(GET /admin/role/:id/perm)
//
// 即本接口提供的是"有哪些可选",而 role_perm 提供的是"已经选了哪些"。
// 两者的交集决定复选框的勾选状态。
//
// 这也解释了为什么 menu_perm 变了不影响判权 ——
// 它只影响第 ② 步展示什么,判权只看 role_perm。
//
// 响应是 {list, total},元素是权限(不是"菜单-权限关联记录")。
func (l *ListMenuPermsLogic) ListMenuPerms(req *types.RelIdReq) (*types.ListMenuPermsResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListMenuPerms(ctx, &v1_userv1.ListMenuPermsReq{
		MenuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListMenuPermsResp{
		List:  converter.PermissionItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
