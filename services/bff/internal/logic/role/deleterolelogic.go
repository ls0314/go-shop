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

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteRole 删除角色(管理端)。
//
// ============================================================
// 删角色的边界由服务端决定
// ============================================================
//
// 几类"该不该拒绝"的判断全部属于 user-service 的领域知识:
//
//	is_system 角色不可删   —— proto 注释写明"系统内置角色不可改、不可删",
//	                          但 BFF 看不到 is_system(要先查一次,
//	                          那就成了 TOCTOU:查完到删之间可能被改)
//	角色还绑着用户?        —— 需要看 user_role 表,那是服务端的库
//	角色还绑着权限/菜单?   —— 同理,服务端的关联表
//
// 故 BFF 只转发。任何在这里"顺手加"的检查都会:
//
//	要么多一次 RPC(且引入 TOCTOU)
//	要么与真正的权威漂移
//
// 服务端会因为"系统角色不可删"返回 error_msg,那时 BFF 回 400 ——
// 前端拿到的是准确原因。
func (l *DeleteRoleLogic) DeleteRole(req *types.RoleIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.DeleteRole(ctx, &v1_userv1.DeleteRoleReq{
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 / 系统内置角色不可删 → 400
		return nil, err
	}

	// 单体: utils.Success(c, nil) → "data": null
	// 这里回 &types.Empty{} → "data": {}。这是 .api 里 returns (Empty)
	// 的统一代价(见 deleteuserinfologic.go 的说明)。
	return &types.Empty{}, nil
}
