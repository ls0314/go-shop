package scopeservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScopeLogic {
	return &CreateScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateScopeLogic) CreateScope(in *v1_userv1.CreateScopeReq) (*v1_userv1.CreateScopeResp, error) {
	scope := converter.FromProtoScope(in.Scope)

	// 关联角色必须存在
	if _, err := l.svcCtx.RoleRepo.GetRoleById(scope.RoleId); err != nil {
		return &v1_userv1.CreateScopeResp{ErrorMsg: err.Error()}, nil
	}

	// 角色 + 资源类型 + 字段名 联合唯一
	existing, err := l.svcCtx.ScopeRepo.GetScopeByUnique(scope.RoleId, scope.ResourceType, scope.FieldName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &v1_userv1.CreateScopeResp{ErrorMsg: model.ScopeExist.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.ScopeRepo.WithTx(tx).CreateScope(scope)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.CreateScopeResp{Scope: converter.ToProtoScope(scope)}, nil
}
