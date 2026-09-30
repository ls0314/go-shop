package roleservice

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, svc *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		ctx:    ctx,
		svcCtx: svc,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRolesLogic) ListRoles(in *v1_userv1.ListRolesReq) (*v1_userv1.ListRolesResp, error) {
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

	roles, total, err := l.svcCtx.RoleRepo.GetRoleList(page, pageSize, in.RoleType)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Role, 0, len(roles))
	for i := range roles {
		items = append(items, converter.ToProtoRole(&roles[i]))
	}
	return &v1_userv1.ListRolesResp{
		Items: items,
		Total: total,
	}, nil
}
