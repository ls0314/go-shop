package rolepermservicelogic

import (
	"context"
	"fmt"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AssignRolePermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermsLogic {
	return &AssignRolePermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignRolePerms 全量替换角色的权限绑定:先清空再写入。
func (l *AssignRolePermsLogic) AssignRolePerms(in *v1_userv1.AssignRolePermsReq) (*v1_userv1.AssignRolePermsResp, error) {
	if _, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId); err != nil {
		return &v1_userv1.AssignRolePermsResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	// 写入前逐个校验权限存在,避免留下无效绑定
	for _, permId := range in.PermIds {
		if _, err := l.svcCtx.PermRepo.GetPermByID(permId); err != nil {
			return &v1_userv1.AssignRolePermsResp{ErrorMsg: model.PermissionNotExist.Error()}, nil
		}
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repo := l.svcCtx.RolePermRepo.WithTx(tx)
		if err := repo.DeleteByRoleId(in.RoleId); err != nil {
			return err
		}
		return repo.CreateRolePerm(in.RoleId, in.PermIds)
	})
	if err != nil {
		return nil, err
	}

	l.invalidateHolderCache(in.RoleId)
	return &v1_userv1.AssignRolePermsResp{}, nil
}

// invalidateHolderCache 清掉该角色下所有用户的权限码缓存。
func (l *AssignRolePermsLogic) invalidateHolderCache(roleId int64) {
	invalidateRoleHolderCache(l.svcCtx, l.Logger, roleId)
}

// invalidateRoleHolderCache 清掉该角色下所有用户的权限码缓存。
// 角色权限变了,持有者的权限集合随之过期;失败只记日志,
// 已提交的绑定不应因为缓存没清掉而报错。
func invalidateRoleHolderCache(svcCtx *svc.ServiceContext, logger logx.Logger, roleId int64) {
	userIds, err := svcCtx.UserRoleRepo.ListUserIdsByRoleId(roleId)
	if err != nil {
		logger.Errorf("查询角色持有者失败,权限缓存未失效: roleId=%d err=%v", roleId, err)
		return
	}
	keys := make([]string, 0, len(userIds))
	for _, uid := range userIds {
		keys = append(keys, fmt.Sprintf("user:perm:%d", uid))
	}
	if len(keys) == 0 {
		return
	}
	if _, err := svcCtx.Redis.Del(keys...); err != nil {
		logger.Errorf("失效权限缓存失败: roleId=%d keys=%v err=%v", roleId, keys, err)
	}
}

type ListRolePermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolePermsLogic {
	return &ListRolePermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRolePermsLogic) ListRolePerms(in *v1_userv1.ListRolePermsReq) (*v1_userv1.ListRolePermsResp, error) {
	permIds, err := l.svcCtx.RolePermRepo.ListPermIdsByRoleId(in.RoleId)
	if err != nil {
		return nil, err
	}
	if len(permIds) == 0 {
		return &v1_userv1.ListRolePermsResp{Items: []*v1_userv1.Permission{}}, nil
	}

	perms, err := l.svcCtx.PermRepo.ListPermsByIds(permIds)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Permission, 0, len(perms))
	for _, p := range perms {
		items = append(items, converter.ToProtoPermission(p))
	}
	return &v1_userv1.ListRolePermsResp{
		Items: items,
		Total: int64(len(items)),
	}, nil
}

type ClearRolePermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearRolePermsLogic {
	return &ClearRolePermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearRolePermsLogic) ClearRolePerms(in *v1_userv1.ClearRolePermsReq) (*v1_userv1.ClearRolePermsResp, error) {
	if _, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId); err != nil {
		return &v1_userv1.ClearRolePermsResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.RolePermRepo.WithTx(tx).DeleteByRoleId(in.RoleId)
	})
	if err != nil {
		return nil, err
	}

	// 清空后该角色不再授予任何权限,持有者缓存同样过期
	invalidateRoleHolderCache(l.svcCtx, l.Logger, in.RoleId)
	return &v1_userv1.ClearRolePermsResp{}, nil
}
