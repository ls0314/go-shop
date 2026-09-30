package scopeservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScopeLogic {
	return &DeleteScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteScopeLogic) DeleteScope(in *v1_userv1.DeleteScopeReq) (*v1_userv1.DeleteScopeResp, error) {
	existing, err := l.svcCtx.ScopeRepo.GetScopeById(in.ScopeId)
	if err != nil || existing == nil {
		return &v1_userv1.DeleteScopeResp{ErrorMsg: model.ScopeNotExist.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.ScopeRepo.WithTx(tx).DeleteScope(in.ScopeId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.DeleteScopeResp{}, nil
}
