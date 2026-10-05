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

type DeleteMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuLogic {
	return &DeleteMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteMenu 删除菜单(管理端)。
//
// ============================================================
// 有子菜单时能不能删?由服务端决定
// ============================================================
//
// 两种合理策略:
//
//	拒绝删除(返回 error_msg)      → 前端要先删子节点
//	级联删除(连子菜单一起删)      → 一次点击删掉整棵子树
//
// **BFF 不做预检**:预检要先查子节点,那是服务端的库;
// 而且"查完到删之间别人加了子节点"是 TOCTOU。
//
// 服务端返回的 error_msg 会经 BFF 回 400,前端拿到准确原因。
// 若实际是级联删除,前端要给出足够的确认提示 —— 那是前端的事,
// BFF 不替它决定。
//
// ============================================================
// 删菜单会影响动态路由
// ============================================================
//
// 菜单是前端动态路由的数据源(GetMenuTreeByUserId)。删掉一个菜单后:
//
//	该路由消失 → 用户直接访问那个 URL 会 404(前端路由没注册)
//	但它对应的**后端接口仍然存在** → 权限体系里那些权限点可能是
//	                                 menu 类型,删了菜单后它们
//	                                 变成"没有入口的权限"
//
// 故删菜单通常要同时清理相关的 menu_perm 绑定。服务端是否级联
// 清掉那些绑定,需要核实 —— 若不清,会留下孤儿绑定。
//
// 核对方法:删一个菜单后查 menu_perm 表,看它那几条还在不在。
func (l *DeleteMenuLogic) DeleteMenu(req *types.MenuIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.DeleteMenu(ctx, &v1_userv1.DeleteMenuReq{
		MenuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 菜单不存在 / 有子菜单不可删 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
