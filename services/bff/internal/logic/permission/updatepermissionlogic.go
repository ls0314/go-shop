// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePermissionLogic {
	return &UpdatePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdatePermission 局部更新权限点(管理端)。
//
// ============================================================
// 可改字段只有 3 个,而且**不含 permission_code**
// ============================================================
//
// .api 里 UpdatePermissionReq 是 permission_name / api_path /
// description 三个指针。**刻意没有 permission_code**:
//
//	permission_code 是判权的键 —— 它出现在 role_perm 的绑定里、
//	也可能被前端的按钮级权限判断硬编码。改它等于
//	"把所有已绑定它的角色悄悄换成了另一个权限"。
//
// 要改编码的正确做法是删了重建(那时绑定关系会显式断掉,
// 能看出影响了谁),而不是就地改字符串。
//
// 同理 **permission_type 也不可改**:api ↔ menu ↔ button 之间的
// 切换会让 request_method / api_path 这两个字段的意义反转
// (它们是 api 类型专有的)。
//
// ============================================================
// api_path 可改,但要小心
// ============================================================
//
// 改 api_path 会让该权限点与路由的对应关系变化 ——
// 若两处不一致,那条接口的判权会失效(对谁都 403,或对谁都放行
// 取决于服务端的匹配方式)。
//
// **BFF 不校验一致性**:它看不到路由注册表与权限数据的全貌,
// 做不了这个判断。这属于"改完之后要验证"的运维动作。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// 全部为 nil 时回 400(见 errNothingToUpdate)。
func (l *UpdatePermissionLogic) UpdatePermission(req *types.UpdatePermissionReq) (*types.UpdatePermissionResp, error) {
	updates := compactFields(
		strField("permission_name", req.PermissionName),
		strField("api_path", req.ApiPath),
		strField("description", req.Description),
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.UpdatePermission(ctx, &v1_userv1.UpdatePermissionReq{
		PermissionId: req.Id,
		Updates:      updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 权限不存在 / 试图改系统内置权限 → 400
		return nil, err
	}

	// proto 注释:UpdatePermissionResp.permission 返回**合并后的完整对象**
	return &types.UpdatePermissionResp{
		PermissionItem: converter.PermissionItem(resp.GetPermission()),
	}, nil
}
