// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRoleMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRoleMenusLogic {
	return &ListRoleMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListRoleMenus 取某角色**已绑定**的菜单。
//
// 路径参数 :id 的语义是 role_id(与 ListRolePerms 同理)。
//
// ============================================================
// 与 GetMenuTreeByRoleId 的区别 —— 两条路径里都有 role id
// ============================================================
//
//	GET /admin/role/:id/menu    ← 本接口。已绑定的菜单(**扁平**列表)
//	GET /admin/menu/:id/tree    ← **全部**菜单的树(递归,带 children)
//
// 前者答"这个角色现在有什么",后者答"菜单全集长什么样"。
// 角色分配菜单页通常两个都要:用后者渲染整棵树,
// 用前者决定哪些节点打勾。
//
// 响应形状也不同:
//
//	本接口  → {list, total}
//	菜单树  → 裸数组(递归)
//
// 不要因为"都跟 role 有关"就合并或改形状。
//
// ============================================================
// 返回的菜单是扁平的,children 为空
// ============================================================
//
// proto 的 Menu.children 在列表查询时为空(注释写明:
// "children 递归,树查询时填充,列表查询时为空")。
//
// 故这里的 converter.MenuItems 会为每个元素生成 children: []
// —— 前端按扁平列表渲染打勾状态即可,不需要层级。
func (l *ListRoleMenusLogic) ListRoleMenus(req *types.RelIdReq) (*types.ListRoleMenusResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListRoleMenus(ctx, &v1_userv1.ListRoleMenusReq{
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListRoleMenusResp{
		List:  converter.MenuItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
