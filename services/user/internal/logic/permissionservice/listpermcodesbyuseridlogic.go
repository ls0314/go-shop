package permissionservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermCodesByUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermCodesByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermCodesByUserIdLogic {
	return &ListPermCodesByUserIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPermCodesByUserIdLogic) ListPermCodesByUserId(in *v1_userv1.ListPermCodesByUserIdReq) (*v1_userv1.ListPermCodesByUserIdResp, error) {
	var codes []string
	err := l.svcCtx.DB.Table("sys_user_role AS ur").
		Joins("JOIN sys_role_permission AS rp ON ur.role_id = rp.role_id").
		Joins("JOIN sys_permission AS p ON rp.permission_id = p.permission_id").
		Where("ur.user_id = ?", in.UserId).
		Distinct("p.permission_code").
		Pluck("p.permission_code", &codes).Error
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ListPermCodesByUserIdResp{
		PermCodes: codes,
	}, nil
}
