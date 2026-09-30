package scopeservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScopeLogic {
	return &GetScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetScopeLogic) GetScope(in *v1_userv1.GetScopeReq) (*v1_userv1.GetScopeResp, error) {
	scope, err := l.svcCtx.ScopeRepo.GetScopeById(in.ScopeId)
	if err != nil {
		// 业务失败:error 返回 nil,原因放 error_msg
		return &v1_userv1.GetScopeResp{ErrorMsg: model.ScopeNotExist.Error()}, nil
	}

	return &v1_userv1.GetScopeResp{Scope: converter.ToProtoScope(scope)}, nil
}
