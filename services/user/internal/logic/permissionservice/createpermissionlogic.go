package permissionservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreatePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePermissionLogic {
	return &CreatePermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePermissionLogic) CreatePermission(in *v1_userv1.CreatePermissionReq) (*v1_userv1.CreatePermissionResp, error) {
	perm := &model.SysPermission{
		PermissionCode: in.Permission.GetPermissionCode(),
		PermissionName: in.Permission.GetPermissionName(),
		PermissionType: in.Permission.GetPermissionType(),
		RequestMethod:  in.Permission.GetRequestMethod(),
		ApiPath:        in.Permission.GetApiPath(),
		Description:    in.Permission.GetDescription(),
		IsSystem:       in.Permission.GetIsSystem(),
	}
	existing, _ := l.svcCtx.PermRepo.GetPermByCodeUk(perm.ApiPath, perm.RequestMethod, perm.PermissionCode)
	if existing != nil {
		return &v1_userv1.CreatePermissionResp{ErrorMsg: model.PermissionExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.PermRepo.WithTx(tx).CreatePerm(perm)
	})
	if err != nil {
		return nil, err
	}

	if _, err := l.svcCtx.Redis.Incr("api:perm:version"); err != nil {
		l.Errorf("递增权限版本号失败：%v", err)
	}
	return &v1_userv1.CreatePermissionResp{Permission: converter.ToProtoPermission(perm)}, nil

}
