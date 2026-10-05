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

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListRoles 分页查询角色(管理端)。
//
// ============================================================
// 响应是 {list, total},不是 proto 的 {items, total}
// ============================================================
//
// 单体 GetRoleList:
//
//	utils.Success(c, gin.H{
//		"list":  roleList,
//		"total": total,
//	})
//
// handler 手工把 proto 的 resp.Items 改名成了 list。
// **12 个列表接口全部如此** —— 故 .api 里所有 List*Resp 都是
// `List []X `json:"list"“。
//
// 这与 ListUsers 的区别只在**分页参数名**:
//
//	GET /admin/user       query: page, pageSize(camelCase)
//	GET /admin/role       query: page, page_size(snake_case)
//
// 都是既有的不一致,不要统一。
//
// ============================================================
// 分页兜底
// ============================================================
//
// 与 ListUsers 同一问题:.api 的 optional int 不传时为 0,
// 而 proto 传 0 会让服务端行为不确定。故兜成 1/10。
func (l *ListRolesLogic) ListRoles(req *types.ListRolesReq) (*types.ListRolesResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListRoles(ctx, &v1_userv1.ListRolesReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		RoleType: req.RoleType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListRolesResp{
		List:  converter.RoleItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
