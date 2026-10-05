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

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetRole 根据 ID 查角色(管理端)。
//
// ============================================================
// 响应是**裸角色对象**,没有 role 那一层包装
// ============================================================
//
// 单体: utils.Success(c, role) —— data 就是那个角色:
//
//	{ "code":200, "message":"Success", "data": { "role_id":1, ... } }
//
// 而 proto 的 GetRoleResp 是 { Role role = 1; string error_msg = 2; } ——
// **不能照 proto 写 .api**,否则前端解 res.data.role_name 会拿到
// undefined(值在 res.data.role.role_name)。
//
// 故 .api 里 GetRoleResp 是**匿名内嵌 RoleItem**:
//
//	GetRoleResp { RoleItem }   → 序列化后字段提升,与裸对象一致
//
// 内嵌的代价:*RoleItem 与 *GetRoleResp 是两个具名类型,
// 不能直接赋值,要先取出来再拷进内嵌字段。
func (l *GetRoleLogic) GetRole(req *types.RoleIdReq) (*types.GetRoleResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetRole(ctx, &v1_userv1.GetRoleReq{
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 → 400
		return nil, err
	}

	item := converter.RoleItem(resp.GetRole())
	return &types.GetRoleResp{RoleItem: item}, nil
}
