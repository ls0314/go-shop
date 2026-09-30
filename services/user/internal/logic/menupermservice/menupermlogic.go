package menupermservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AssignMenuPermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignMenuPermsLogic {
	return &AssignMenuPermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignMenuPerms 全量替换菜单的权限绑定:先清空再写入。
func (l *AssignMenuPermsLogic) AssignMenuPerms(in *v1_userv1.AssignMenuPermsReq) (*v1_userv1.AssignMenuPermsResp, error) {
	if _, err := l.svcCtx.MenuRepo.GetMenuById(in.MenuId); err != nil {
		return &v1_userv1.AssignMenuPermsResp{ErrorMsg: model.MenuNotExist.Error()}, nil
	}

	for _, permId := range in.PermIds {
		if _, err := l.svcCtx.PermRepo.GetPermByID(permId); err != nil {
			return &v1_userv1.AssignMenuPermsResp{ErrorMsg: model.PermissionNotExist.Error()}, nil
		}
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repo := l.svcCtx.MenuPermRepo.WithTx(tx)
		if err := repo.DeleteByMenuId(in.MenuId); err != nil {
			return err
		}
		return repo.CreateMenuPerm(in.MenuId, in.PermIds)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.AssignMenuPermsResp{}, nil
}

type ListMenuPermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMenuPermsLogic {
	return &ListMenuPermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMenuPermsLogic) ListMenuPerms(in *v1_userv1.ListMenuPermsReq) (*v1_userv1.ListMenuPermsResp, error) {
	permIds, err := l.svcCtx.MenuPermRepo.ListPermIdsByMenuId(in.MenuId)
	if err != nil {
		return nil, err
	}
	if len(permIds) == 0 {
		return &v1_userv1.ListMenuPermsResp{Items: []*v1_userv1.Permission{}}, nil
	}

	perms, err := l.svcCtx.PermRepo.ListPermsByIds(permIds)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Permission, 0, len(perms))
	for _, p := range perms {
		items = append(items, converter.ToProtoPermission(p))
	}
	return &v1_userv1.ListMenuPermsResp{
		Items: items,
		Total: int64(len(items)),
	}, nil
}

type ClearMenuPermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearMenuPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearMenuPermsLogic {
	return &ClearMenuPermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearMenuPermsLogic) ClearMenuPerms(in *v1_userv1.ClearMenuPermsReq) (*v1_userv1.ClearMenuPermsResp, error) {
	if _, err := l.svcCtx.MenuRepo.GetMenuById(in.MenuId); err != nil {
		return &v1_userv1.ClearMenuPermsResp{ErrorMsg: model.MenuNotExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.MenuPermRepo.WithTx(tx).DeleteByMenuId(in.MenuId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ClearMenuPermsResp{}, nil
}
