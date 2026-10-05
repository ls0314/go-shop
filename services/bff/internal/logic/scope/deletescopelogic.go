// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package scope

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScopeLogic {
	return &DeleteScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteScope 删除数据权限规则(管理端)。
//
// ============================================================
// 删规则的后果:该角色的可见数据范围**变大**
// ============================================================
//
// 数据权限是**收窄**可见范围的(WHERE 条件),故删掉一条规则
// 意味着那个角色**看到的行变多** —— 方向与"删菜单则菜单消失"
// 相反,而且不会报错、不会有提示。
//
// 具体到一条规则 `dept_id IN (1,2,3)`:
//
//	删掉它后,若该角色没有其它收窄规则 → 它可能看到**全部部门**的数据
//
// 这是最典型的"权限放大"事故形态:一次删除操作让某角色
// 突然能看全站数据,而系统没有任何提示。
//
// **BFF 不做拦截**(它无法判断"删了之后还剩什么约束")——
// 但这条接口在运维上属于高危。前端的确认弹窗应当明确提示
// "删除后该角色的可见数据范围可能扩大"。
//
// 服务端若实现了"每个角色至少一条规则"之类的保护,会返回
// error_msg,BFF 回 400。
func (l *DeleteScopeLogic) DeleteScope(req *types.ScopeIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.DeleteScope(ctx, &v1_userv1.DeleteScopeReq{
		ScopeId: req.Id,
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
