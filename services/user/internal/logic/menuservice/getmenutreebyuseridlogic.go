package menuservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeByUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuTreeByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeByUserIdLogic {
	return &GetMenuTreeByUserIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuTreeByUserIdLogic) GetMenuTreeByUserId(in *v1_userv1.GetMenuTreeByUserIdReq) (*v1_userv1.GetMenuTreeByUserIdResp, error) {
	roleIds, err := l.svcCtx.MenuRepo.ListRoleIdsByUserId(in.GetUserId())
	if err != nil {
		return nil, err
	}

	menuIds, err := l.svcCtx.MenuRepo.ListMenuIdsByRoleIds(roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return &v1_userv1.GetMenuTreeByUserIdResp{Items: []*v1_userv1.Menu{}}, nil
	}

	menuList, err := l.svcCtx.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, err
	}

	return &v1_userv1.GetMenuTreeByUserIdResp{
		Items: converter.ToProtoMenuTree(MakeTree(menuList)),
	}, nil
}
