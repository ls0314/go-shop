package roleservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRoleLogic) UpdateRole(in *v1_userv1.UpdateRoleReq) (*v1_userv1.UpdateRoleResp, error) {
	oldRole, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId)
	if err != nil {
		return &v1_userv1.UpdateRoleResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	if oldRole.IsSystem {
		return &v1_userv1.UpdateRoleResp{ErrorMsg: model.RoleIsSystem.Error()}, nil
	}

	newRole := *oldRole
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newRole,
	})

	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateRoleResp{ErrorMsg: err.Error()}, nil
	}

	if newRole.RoleName != oldRole.RoleName {
		existing, _ := l.svcCtx.RoleRepo.GetRoleByName(newRole.RoleName)
		if existing != nil {
			return &v1_userv1.UpdateRoleResp{ErrorMsg: model.RoleExist.Error()}, nil
		}
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.RoleRepo.WithTx(tx).UpdateRole(&newRole)
	})

	if err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateRoleResp{Role: converter.ToProtoRole(&newRole)}, nil
}
