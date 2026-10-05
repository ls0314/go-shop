// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package scope

import (
	"context"

	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateScopeLogic {
	return &UpdateScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateScope 局部更新数据权限规则(管理端)。
//
// ============================================================
// 可改字段里没有 role_id —— 这是刻意的
// ============================================================
//
// .api 里 UpdateScopeReq 是 resource_type / field_name /
// condition_type / condition_value / description 五个指针,
// **没有 role_id**。
//
// 理由:改 role_id 等于"把这条规则从 A 角色挪到 B 角色" ——
// 那不是更新,是移动。语义上应当"在 B 下建一条、把 A 的删掉",
// 或者由服务端提供一个显式的 transfer 接口。
//
// 就地改 role_id 的后果很隐蔽:A 角色突然失去一条数据权限、
// B 角色突然多一条,而审计日志里只有一条 update。故不暴露它。
//
// ============================================================
// 改 condition_* 是安全敏感操作
// ============================================================
//
// 这几个字段会被拼进 SQL 的 WHERE 子句(见 createscopelogic.go
// 的说明)。把它们从"受控值"改成另一个值,可能让某个角色的
// **可见数据范围突然扩大** —— 而这不是一个会报错的操作。
//
// BFF 不做额外校验(不知道服务端的拼接方式),
// 但这条接口在运维上属于"改完要验证数据范围"的操作。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// 全部为 nil 时回 400(errNothingToUpdate)。
func (l *UpdateScopeLogic) UpdateScope(req *types.UpdateScopeReq) (*types.UpdateScopeResp, error) {
	updates := compactFields(
		strField("resource_type", req.ResourceType),
		strField("field_name", req.FieldName),
		strField("condition_type", req.ConditionType),
		strField("condition_value", req.ConditionValue),
		strField("description", req.Description),
		// role_id 不在列 —— 见上方说明
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.UpdateScope(ctx, &v1_userv1.UpdateScopeReq{
		ScopeId: req.Id,
		Updates: updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// proto 注释:UpdateScopeResp.scope 返回**合并后的完整对象**
	return &types.UpdateScopeResp{
		ScopeItem: converter.ScopeItem(resp.GetScope()),
	}, nil
}
