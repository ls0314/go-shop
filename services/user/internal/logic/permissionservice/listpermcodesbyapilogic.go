package permissionservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermCodesByApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermCodesByApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermCodesByApiLogic {
	return &ListPermCodesByApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPermCodesByApiLogic) ListPermCodesByApi(in *v1_userv1.ListPermCodesByApiReq) (*v1_userv1.ListPermCodesByApiResp, error) {
	var codes []string
	err := l.svcCtx.DB.Table("sys_permission").
		Where("api_path = ? AND request_method = ?", in.ApiPath, in.RequestMethod).
		Pluck("permission_code", &codes).Error
	if err != nil {
		return nil, err
	}
	return &v1_userv1.ListPermCodesByApiResp{
		PermCodes: codes,
	}, nil
}
