// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUserRolesLogic {
	return &ClearUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClearUserRoles 清空某用户的角色绑定。
//
// ============================================================
// 这个方法与 AssignUserRoles 传空数组**功能重叠,但保留两者**
// ============================================================
//
//	AssignUserRoles(userId, [])   ← 也能清空
//	ClearUserRoles(userId)         ← 本接口
//
// 保留的理由:
//
//	① 单体两条路由都在(前端可能各有用到的地方);
//	② 语义不同:前者是"把集合设成空",后者是"删除全部绑定"——
//	   在服务端实现上可能是 upsert vs delete,审计日志的措辞也不同;
//	③ 删掉一条会破坏前端契约,而这是迁移期最不能做的事。
//
// BFF 只做转发,不在这里做"去重"(比如把 clear 翻译成
// assign-role with []) —— 那会让审计日志记错动作。
//
// ============================================================
// 幂等
// ============================================================
//
// 对已经无绑定的用户再清一次应当成功(不报"没有绑定可清")。
// 这是服务端的责任,BFF 不预检 —— 预检会引入 TOCTOU:
// 检查时没绑定,删的时候别人加上了,于是那次删除白做但报成功。
func (l *ClearUserRolesLogic) ClearUserRoles(req *types.RelIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ClearUserRoles(ctx, &v1_userv1.ClearUserRolesReq{
		UserId: req.Id,
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
