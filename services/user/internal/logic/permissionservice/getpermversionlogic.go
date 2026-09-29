package permissionservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"
	"strconv"

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
	v, err := l.svcCtx.Redis.Get("api:perm:version")
	if err != nil {
		return &v1_userv1.GetPermVersionResp{Version: 0}, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return &v1_userv1.GetPermVersionResp{Version: 0}, nil
	}
	return &v1_userv1.GetPermVersionResp{Version: n}, nil
}
