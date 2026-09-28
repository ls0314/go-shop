package permissionservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPermVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermVersionLogic {
	return &GetPermVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPermVersionLogic) GetPermVersion(in *v1_userv1.GetPermVersionReq) (*v1_userv1.GetPermVersionResp, error) {
	return &v1_userv1.GetPermVersionResp{Version: 0}, nil
}
