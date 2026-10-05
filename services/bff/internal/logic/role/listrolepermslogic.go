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

type ListRolePermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolePermsLogic {
	return &ListRolePermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListRolePerms 取某角色已绑定的权限(角色分配权限页回显已选)。
//
// ============================================================
// 路径参数 :id 的语义是 role_id
// ============================================================
//
// 路由 GET /api/v1/admin/role/:id/perm → RPC 的 role_id。
// .api 里沿用 :id(types.RelIdReq),故这里显式映射 req.Id → RoleId。
//
// 同理 ListRoleMenus 也是 :id → role_id。
//
// ============================================================
// 响应是 {list, total},元素是权限
// ============================================================
//
// 单体 role_perm_handler.go:
//
//	utils.Success(c, gin.H{
//		"list":  rolePermList,
//		"total": total,
//	})
//
// 元素类型是**权限**(不是"角色-权限关联记录")——
// 前端拿到的是权限列表,用来在权限树上打勾。
// 故 converter.PermissionItems,不是某个 RelItem 类型。
//
// ============================================================
// 没有分页参数
// ============================================================
//
// 这条没有 page/page_size —— 一个角色绑定的权限数量有限
// (通常几十到几百),全量返回即可。total 是元素个数。
func (l *ListRolePermsLogic) ListRolePerms(req *types.RelIdReq) (*types.ListRolePermsResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListRolePerms(ctx, &v1_userv1.ListRolePermsReq{
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListRolePermsResp{
		List:  converter.PermissionItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
