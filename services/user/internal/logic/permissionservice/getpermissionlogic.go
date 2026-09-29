package permissionservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionLogic {
	return &GetPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPermissionLogic) GetPermission(in *v1_userv1.GetPermissionReq) (*v1_userv1.GetPermissionResp, error) {
	perm, err := l.svcCtx.PermRepo.GetPermByID(in.PermissionId)
	if err != nil {
		return &v1_userv1.GetPermissionResp{ErrorMsg: model.PermissionExist.Error()}, nil
	}
	return &v1_userv1.GetPermissionResp{Permission: converter.ToProtoPermission(perm)}, nil
}
