package menuservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeByRoleIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuTreeByRoleIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeByRoleIdLogic {
	return &GetMenuTreeByRoleIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuTreeByRoleIdLogic) GetMenuTreeByRoleId(in *v1_userv1.GetMenuTreeByRoleIdReq) (*v1_userv1.GetMenuTreeByRoleIdResp, error) {
	menuIds, err := l.svcCtx.MenuRepo.ListMenuIdsByRoleIds([]int64{in.RoleId})
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return &v1_userv1.GetMenuTreeByRoleIdResp{Items: []*v1_userv1.Menu{}}, nil
	}

	menuList, err := l.svcCtx.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, err
	}

	tree, warnings := MakeTree(menuList)
	if len(warnings) > 0 {
		l.Errorf("菜单树数据异常(role_id=%d): %s", in.RoleId, FormatTreeWarnings(warnings))
	}

	return &v1_userv1.GetMenuTreeByRoleIdResp{
		Items: converter.ToProtoMenuTree(tree),
	}, nil
}
