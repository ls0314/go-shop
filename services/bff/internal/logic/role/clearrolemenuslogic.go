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

type ClearRoleMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearRoleMenusLogic {
	return &ClearRoleMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClearRoleMenus 清空某角色的菜单绑定。
//
// ============================================================
// 与 AssignRoleMenus 传空数组**功能重叠,但保留两者**
// ============================================================
//
//	AssignRoleMenus(roleId, [])   ← 也能清空
//	ClearRoleMenus(roleId)         ← 本接口
//
// 保留的理由:
//
//	① 单体两条路由都在,删一条会破坏前端契约(迁移期最不能做的事);
//	② 服务端实现可能不同(delete-all vs upsert-empty),
//	   审计日志的措辞也不同;
//	③ 语义更直白 —— "清空"不需要前端先想清楚"空数组会不会被当成没传"。
//
// BFF 不在这里做"去重"(比如把 clear 翻译成 assign with [])——
// 那会让审计日志记错动作。
//
// ============================================================
// 幂等
// ============================================================
//
// 对已经无绑定的角色再清一次应当成功。BFF 不预检 ——
// 预检会引入 TOCTOU:检查时没绑定,删的时候别人加上了,
// 于是那次删除白做但报成功。
func (l *ClearRoleMenusLogic) ClearRoleMenus(req *types.RelIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ClearRoleMenus(ctx, &v1_userv1.ClearRoleMenusReq{
		RoleId: req.Id,
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
