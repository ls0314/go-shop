package userroleservicelogic

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

type AssignUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRolesLogic {
	return &AssignUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignUserRoles 全量替换用户的角色绑定:先清空再写入。
func (l *AssignUserRolesLogic) AssignUserRoles(in *v1_userv1.AssignUserRolesReq) (*v1_userv1.AssignUserRolesResp, error) {
	if _, err := l.svcCtx.UserRepo.GetUserById(in.UserId); err != nil {
		return &v1_userv1.AssignUserRolesResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	for _, roleId := range in.RoleIds {
		if _, err := l.svcCtx.RoleRepo.GetRoleById(roleId); err != nil {
			return &v1_userv1.AssignUserRolesResp{ErrorMsg: model.RoleNotExist.Error()}, nil
		}
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repo := l.svcCtx.UserRoleRepo.WithTx(tx)
		if err := repo.DeleteByUserId(in.UserId); err != nil {
			return err
		}
		return repo.CreateUserRole(in.UserId, in.RoleIds)
	})
	if err != nil {
		return nil, err
	}

	invalidateUserPermCache(l.svcCtx, l.Logger, in.UserId)
	return &v1_userv1.AssignUserRolesResp{}, nil
}

// invalidateUserPermCache 清掉该用户的权限码缓存。
// 角色绑定变了,用户的权限集合随之过期;失败只记日志,
// 已提交的绑定不应因为缓存没清掉而报错。
func invalidateUserPermCache(svcCtx *svc.ServiceContext, logger logx.Logger, userId int64) {
	key := fmt.Sprintf("user:perm:%d", userId)
	if _, err := svcCtx.Redis.Del(key); err != nil {
		logger.Errorf("失效权限缓存失败: userId=%d err=%v", userId, err)
	}
}

type ListUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRolesLogic {
	return &ListUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserRolesLogic) ListUserRoles(in *v1_userv1.ListUserRolesReq) (*v1_userv1.ListUserRolesResp, error) {
	roleIds, err := l.svcCtx.UserRoleRepo.ListRoleIdsBySingleUserId(in.UserId)
	if err != nil {
		return nil, err
	}
	if len(roleIds) == 0 {
		return &v1_userv1.ListUserRolesResp{Items: []*v1_userv1.Role{}}, nil
	}

	roles, err := l.svcCtx.RoleRepo.ListRolesByIds(roleIds)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Role, 0, len(roles))
	for _, r := range roles {
		items = append(items, converter.ToProtoRole(r))
	}
	return &v1_userv1.ListUserRolesResp{
		Items: items,
		Total: int64(len(items)),
	}, nil
}

type ClearUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUserRolesLogic {
	return &ClearUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearUserRolesLogic) ClearUserRoles(in *v1_userv1.ClearUserRolesReq) (*v1_userv1.ClearUserRolesResp, error) {
	if _, err := l.svcCtx.UserRepo.GetUserById(in.UserId); err != nil {
		return &v1_userv1.ClearUserRolesResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.UserRoleRepo.WithTx(tx).DeleteByUserId(in.UserId)
	})
	if err != nil {
		return nil, err
	}

	invalidateUserPermCache(l.svcCtx, l.Logger, in.UserId)
	return &v1_userv1.ClearUserRolesResp{}, nil
}
