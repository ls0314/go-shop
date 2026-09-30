package scopeservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateScopeLogic {
	return &UpdateScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateScopeLogic) UpdateScope(in *v1_userv1.UpdateScopeReq) (*v1_userv1.UpdateScopeResp, error) {
	oldScope, err := l.svcCtx.ScopeRepo.GetScopeById(in.ScopeId)
	if err != nil || oldScope == nil {
		return &v1_userv1.UpdateScopeResp{ErrorMsg: model.ScopeNotExist.Error()}, nil
	}

	// 关联角色必须存在
	if _, err := l.svcCtx.RoleRepo.GetRoleById(oldScope.RoleId); err != nil {
		return &v1_userv1.UpdateScopeResp{ErrorMsg: err.Error()}, nil
	}

	// 局部更新:用 FieldUpdate 列表覆盖旧对象,只覆盖传入的字段
	newScope := *oldScope
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newScope,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateScopeResp{ErrorMsg: err.Error()}, nil
	}

	// 数据权限的归属角色不允许修改
	if oldScope.RoleId != newScope.RoleId {
		return &v1_userv1.UpdateScopeResp{ErrorMsg: model.ScopeIsRole.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.ScopeRepo.WithTx(tx).UpdateScope(&newScope)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateScopeResp{Scope: converter.ToProtoScope(&newScope)}, nil
}
