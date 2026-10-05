// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateRole 局部更新角色(管理端)。
//
// ============================================================
// 可改字段与"不可改字段"
// ============================================================
//
// .api 里 UpdateRoleReq 有 4 个指针字段:role_name / description /
// data_scope / is_default。
//
// **刻意没有 role_type 与 is_system**:
//
//	role_type   platform / seller / system —— 改它等于把角色在
//	           权限模型里的定位换掉,会让既有的 role_perm 绑定
//	           变得语义不明。属于"要改就删了重建"的字段。
//	is_system   系统内置标记 —— proto 注释写明"系统内置角色不可改、不可删"。
//
// 故这里只暴露运营真正会调的 4 个。若前端确实要改 role_type,
// 那是契约变更,需要单独讨论(而不是在 BFF"顺手"加个字段)。
//
// ============================================================
// is_default 的语义要注意
// ============================================================
//
// proto 注释:"新用户是否自动分配该角色"。
//
// 所以把一个角色设成 is_default=true 会影响**将来**注册的用户,
// 不改动已注册用户。前端若把它理解成"当前用户的默认角色"就错了。
// 这是服务端语义,BFF 只透传 —— 但注释在此提醒。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// 全部为 nil 时回 400(见 errNothingToUpdate 的说明)——
// 这与"客户端传了不认识的字段"是同一种情况。
func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleReq) (*types.UpdateRoleResp, error) {
	updates := compactFields(
		strField("role_name", req.RoleName),
		strField("description", req.Description),
		strField("data_scope", req.DataScope),
		boolField("is_default", req.IsDefault),
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.UpdateRole(ctx, &v1_userv1.UpdateRoleReq{
		RoleId:  req.Id,
		Updates: updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 / 试图改系统内置角色 → 400
		return nil, err
	}

	// proto 注释:UpdateRoleResp.role 返回**合并后的完整对象**,
	// 调用方无需再查一次。
	//
	// 响应是裸对象(内嵌) —— 见 getrolelogic.go 的说明。
	item := converter.RoleItem(resp.GetRole())
	return &types.UpdateRoleResp{RoleItem: item}, nil
}
