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
	return &v1_userv1.ListPermCodesByApiResp{
		PermCodes: []string{"platform:inventory:adjust"},
	}, nil
}
