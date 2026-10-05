// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package scope

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListScopesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListScopesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScopesLogic {
	return &ListScopesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListScopes 分页查询数据权限规则(管理端)。
//
// 响应是 {list, total}(单体 handler 手工包了 gin.H{"list": ...}),
// 不是 proto 的 {items, total}。
//
// ============================================================
// 过滤参数是 resourceType(camelCase)
// ============================================================
//
// 单体 scope_handler.go 用 c.Query("resourceType")。
// 写成 resource_type 会被 httpx.Parse 静默丢弃 →
// 按资源筛选失效但不报错。
//
// ============================================================
// 列表里没有角色名
// ============================================================
//
// proto 的 Scope 只有 role_id,注释写明"角色名由前端用角色列表
// 本地映射,不由本接口返回"。
//
// 故前端渲染这个列表时,要**另外拿一次角色列表**
// (GET /admin/role)建 id→name 的映射表,否则"适用角色"那列
// 只能显示数字 ID。
//
// BFF 不在这里做 join —— 那需要跨域取数并缓存,
// 是前端或后续 BFF 聚合层的职责。
func (l *ListScopesLogic) ListScopes(req *types.ListScopesReq) (*types.ListScopesResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListScopes(ctx, &v1_userv1.ListScopesReq{
		Page:         int32(page),
		PageSize:     int32(pageSize),
		ResourceType: req.ResourceType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListScopesResp{
		List:  converter.ScopeItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
