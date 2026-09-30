package rolemenuservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AssignRoleMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleMenusLogic {
	return &AssignRoleMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignRoleMenus 全量替换角色的菜单绑定:先清空再写入。
func (l *AssignRoleMenusLogic) AssignRoleMenus(in *v1_userv1.AssignRoleMenusReq) (*v1_userv1.AssignRoleMenusResp, error) {
	if _, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId); err != nil {
		return &v1_userv1.AssignRoleMenusResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	for _, menuId := range in.MenuIds {
		if _, err := l.svcCtx.MenuRepo.GetMenuById(menuId); err != nil {
			return &v1_userv1.AssignRoleMenusResp{ErrorMsg: model.MenuNotExist.Error()}, nil
		}
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repo := l.svcCtx.RoleMenuRepo.WithTx(tx)
		if err := repo.DeleteByRoleId(in.RoleId); err != nil {
			return err
		}
		return repo.CreateRoleMenu(in.RoleId, in.MenuIds)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.AssignRoleMenusResp{}, nil
}

type ListRoleMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRoleMenusLogic {
	return &ListRoleMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRoleMenusLogic) ListRoleMenus(in *v1_userv1.ListRoleMenusReq) (*v1_userv1.ListRoleMenusResp, error) {
	menuIds, err := l.svcCtx.RoleMenuRepo.ListMenuIdsBySingleRoleId(in.RoleId)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return &v1_userv1.ListRoleMenusResp{Items: []*v1_userv1.Menu{}}, nil
	}

	menus, err := l.svcCtx.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ListRoleMenusResp{
		Items: converter.ToProtoMenuTree(menus),
		Total: int64(len(menus)),
	}, nil
}

type ClearRoleMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearRoleMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearRoleMenusLogic {
	return &ClearRoleMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearRoleMenusLogic) ClearRoleMenus(in *v1_userv1.ClearRoleMenusReq) (*v1_userv1.ClearRoleMenusResp, error) {
	if _, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId); err != nil {
		return &v1_userv1.ClearRoleMenusResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.RoleMenuRepo.WithTx(tx).DeleteByRoleId(in.RoleId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ClearRoleMenusResp{}, nil
}
