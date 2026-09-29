package permissionservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPermissionsLogic) ListPermissions(in *v1_userv1.ListPermissionsReq) (*v1_userv1.ListPermissionsResp, error) {
	page := int(in.Page)
	if page <= 0 {
		page = 1
	}

	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	perms, total, err := l.svcCtx.PermRepo.GetPermList(page, pageSize, in.PermissionType)
	if err != nil {
		return nil, err
	}
	items := make([]*v1_userv1.Permission, 0, len(perms))
	for i := range perms {
		items = append(items, converter.ToProtoPermission(&perms[i]))
	}
	return &v1_userv1.ListPermissionsResp{Items: items, Total: total}, nil
}
