// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package scope

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScopeLogic {
	return &GetScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetScope 根据 ID 查数据权限规则(管理端)。
//
// 响应是**裸对象**(单体: utils.Success(c, scope)),
// 故 .api 里 GetScopeResp 是匿名内嵌 ScopeItem。
//
// 同样不含角色名(只有 role_id),见 listscopeslogic.go 的说明。
func (l *GetScopeLogic) GetScope(req *types.ScopeIdReq) (*types.GetScopeResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetScope(ctx, &v1_userv1.GetScopeReq{
		ScopeId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 规则不存在 → 400
		return nil, err
	}

	return &types.GetScopeResp{
		ScopeItem: converter.ScopeItem(resp.GetScope()),
	}, nil
}
