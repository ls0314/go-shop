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

type GetMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuLogic {
	return &GetMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetMenu 根据 ID 查菜单(管理端)。
//
// 响应是**裸菜单对象**(单体: utils.Success(c, menu)),
// 没有 menu 那一层包装。故 .api 里 GetMenuResp 是匿名内嵌 MenuItem。
//
// ============================================================
// 拿到的 children 是空的
// ============================================================
//
// 这是单条查询,不是树查询 —— proto 的 children 只在树查询时填充。
// 前端要子节点列表得用 GET /admin/menu/children/:id 之类
// (本项目没有那个接口)或直接用树查询。
//
// 故编辑菜单页通常:用列表/树拿到整棵树 → 定位到目标节点 →
// 渲染表单。而不是"进来先单查一条"。
func (l *GetMenuLogic) GetMenu(req *types.MenuIdReq) (*types.GetMenuResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetMenu(ctx, &v1_userv1.GetMenuReq{
		MenuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 菜单不存在 → 400
		return nil, err
	}

	return &types.GetMenuResp{
		MenuItem: converter.MenuItem(resp.GetMenu()),
	}, nil
}
