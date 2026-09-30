package menuservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenusLogic {
	return &ListMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMenusLogic) ListMenus(in *v1_userv1.ListMenusReq) (*v1_userv1.ListMenusResp, error) {
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

	menus, total, err := l.svcCtx.MenuRepo.GetMenuList(page, pageSize, in.GetMenuType())
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Menu, 0, len(menus))

	for i := range menus {
		items = append(items, converter.ToProtoMenu(&menus[i]))
	}

	return &v1_userv1.ListMenusResp{
		Items: items,
		Total: total,
	}, nil
}
