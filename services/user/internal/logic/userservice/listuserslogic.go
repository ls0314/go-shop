package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUsersLogic) ListUsers(in *v1_userv1.ListUsersReq) (*v1_userv1.ListUsersResp, error) {
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

	users, total, err := l.svcCtx.UserRepo.GetUserList(page, pageSize, in.Status)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.User, 0, len(users))
	for i := range users {
		items = append(items, converter.ToProtoUser(&users[i]))
	}
	return &v1_userv1.ListUsersResp{
		Items: items,
		Total: total,
	}, nil
}
