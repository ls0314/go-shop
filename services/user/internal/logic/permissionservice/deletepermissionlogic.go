package permissionservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeletePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePermissionLogic {
	return &DeletePermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePermissionLogic) DeletePermission(in *v1_userv1.DeletePermissionReq) (*v1_userv1.DeletePermissionResp, error) {
	existing, err := l.svcCtx.PermRepo.GetPermByID(in.PermissionId)
	if err != nil {
		return &v1_userv1.DeletePermissionResp{ErrorMsg: model.PermissionNotExist.Error()}, err
	}

	if existing.IsSystem {
		return &v1_userv1.DeletePermissionResp{ErrorMsg: model.PermissionIsSystem.Error()}, nil
	}

	hasRel, err := l.svcCtx.PermRepo.CheckRoleRelPerm(in.PermissionId)
	if err != nil {
		return nil, err
	}
	if hasRel {
		return &v1_userv1.DeletePermissionResp{ErrorMsg: model.PermissionHasRel.Error()}, nil
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.PermRepo.WithTx(tx).DeletePerm(in.PermissionId)
	})
	if err != nil {
		return nil, err
	}

	if _, err := l.svcCtx.Redis.Incr("api:perm:version"); err != nil {
		l.Errorf("递增权限版本号失败: %v", err)
	}
	return &v1_userv1.DeletePermissionResp{}, nil
}
