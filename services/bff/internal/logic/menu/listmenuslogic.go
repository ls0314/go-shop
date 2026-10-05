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

type ListMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenusLogic {
	return &ListMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListMenus 分页查询菜单(**扁平列表**,不是树)。
//
// 响应是 {list, total},不是 proto 的 {items, total} ——
// 单体 menu_handler.go 手工包了 gin.H{"list": ..., "total": total}。
//
// ============================================================
// 列表里 children 为空,这是正常的
// ============================================================
//
// proto 的 Menu.children 注释写明:"递归,树查询时填充,
// 列表查询时为空"。
//
// 故这里的 converter.MenuItems 会给每个元素生成 children: []
// —— 前端若拿这个列表渲染菜单表格,用**扁平**渲染方式(带
// parent_id 缩进或独立的树形表格组件),不要指望下钻。
//
// 要层级结构用 POST /admin/menu/tree 或 GET /admin/menu/:id/tree。
//
// ============================================================
// 过滤参数是 menuType(camelCase)
// ============================================================
//
// 单体用 c.Query("menuType")。写成 menu_type 会让
// httpx.Parse 静默丢弃 → 按类型筛选失效但不报错。
func (l *ListMenusLogic) ListMenus(req *types.ListMenusReq) (*types.ListMenusResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListMenus(ctx, &v1_userv1.ListMenusReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		MenuType: req.MenuType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListMenusResp{
		List:  converter.MenuItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
