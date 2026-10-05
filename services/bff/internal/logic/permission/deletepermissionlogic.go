// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePermissionLogic {
	return &DeletePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeletePermission 删除权限点(管理端)。
//
// ============================================================
// 删权限的影响面比看起来大
// ============================================================
//
// 一个权限点可能同时被这些引用:
//
//	role_perm     角色-权限绑定(删了它,那些角色少一个权限)
//	menu_perm     菜单-权限绑定(删了它,那些菜单下的按钮不再受控)
//	前端代码      按钮级判断里的权限码字符串(硬编码在前端)
//
// **前两个由服务端处理**(级联清绑定或拒绝删除,取决于它的策略);
// **第三个 BFF 与服务端都看不见** —— 删掉一个权限点后,
// 前端那句 `v-if="perms.includes('platform:coupon:create')"`
// 会永远为 false,而没有任何报错。
//
// 故这条接口在运维上属于"要谨慎"的操作。BFF 不做额外拦截
// (它无法判断前端有没有引用),但**删完之后要回归测试相关页面**。
//
// 服务端会因"系统内置权限不可删"或"仍被引用"返回 error_msg,
// 那时 BFF 回 400,前端拿到准确原因。
func (l *DeletePermissionLogic) DeletePermission(req *types.PermissionIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.DeletePermission(ctx, &v1_userv1.DeletePermissionReq{
		PermissionId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.Empty{}, nil
}
