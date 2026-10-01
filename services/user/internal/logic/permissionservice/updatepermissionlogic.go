package permissionservicelogic

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

type UpdatePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePermissionLogic {
	return &UpdatePermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePermissionLogic) UpdatePermission(in *v1_userv1.UpdatePermissionReq) (*v1_userv1.UpdatePermissionResp, error) {
	olderPerm, err := l.svcCtx.PermRepo.GetPermByID(in.PermissionId)
	if err != nil {
		return &v1_userv1.UpdatePermissionResp{ErrorMsg: model.PermissionNotExist.Error()}, err
	}

	if olderPerm.IsSystem {
		return &v1_userv1.UpdatePermissionResp{ErrorMsg: model.PermissionIsSystem.Error()}, err
	}

	newPerm := *olderPerm
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newPerm,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdatePermissionResp{ErrorMsg: err.Error()}, err
	}
	if newPerm.PermissionCode != olderPerm.PermissionCode {
		return &v1_userv1.UpdatePermissionResp{ErrorMsg: model.PermissionCodeNotAlter.Error()}, err
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.PermRepo.WithTx(tx).UpdatePerm(&newPerm)
	})

	if err != nil {
		return nil, err
	}

	if _, err := l.svcCtx.Redis.Incr("api:perm:version"); err != nil {
		l.Errorf("递增权限版本号失败：%v", err)
	}
	return &v1_userv1.UpdatePermissionResp{Permission: converter.ToProtoPermission(&newPerm)}, nil
}
