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
	return &v1_userv1.ListPermCodesByUserIdResp{
		PermCodes: []string{"platform:inventory:adjust"},
	}, nil
}
