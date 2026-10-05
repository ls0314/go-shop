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

type AssignRoleMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleMenusLogic {
	return &AssignRoleMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssignRoleMenus 全量替换角色的菜单绑定。
//
// 全量替换语义与 AssignRolePerms 完全相同(见那边的详细说明),
// 只是集合换成了菜单 ID:
//
//	menu_ids: []        → 清空该角色的所有菜单(合法输入,不拦)
//	menu_ids: [1,2,3]   → 设成恰好这三个
//
// ============================================================
// 角色-菜单 与 角色-权限 的联动由服务端负责
// ============================================================
//
// 一个容易期望但**不该在 BFF 做**的事:绑了菜单就自动带上
// 该菜单下的权限。
//
// 实际语义是两者独立:菜单决定"侧边栏显示什么",
// 权限决定"按钮能不能点"。一个用户可能有菜单但没有该菜单下的
// 某些按钮权限。所以服务端不会联动,BFF 更不该。
//
// 若前端的交互是"勾了菜单就自动勾上其权限",那是**前端的
// 便利行为**,由它自己发两次请求实现 —— 而不是让后端偷偷联动,
// 否则管理员无法表达"能进这个页面但只能看不能改"。
//
// 响应回显请求体:见 assignrolepermslogic.go 的说明(与单体的已知差异)。
func (l *AssignRoleMenusLogic) AssignRoleMenus(req *types.AssignRoleMenusReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.AssignRoleMenus(ctx, &v1_userv1.AssignRoleMenusReq{
		RoleId:  req.RoleId,
		MenuIds: req.MenuIds,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 / 菜单 ID 不存在 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
