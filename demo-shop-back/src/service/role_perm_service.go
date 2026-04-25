package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"gorm.io/gorm"
)

// RolePermService 角色-权限关联服务层实例
type RolePermService struct {
	RolePermRepo *repository.RolePermRepo   // 角色-权限关联表数据层实例
	RoleRepo     *repository.RoleRepo       // 角色表数据层实例
	PermRepo     *repository.PermissionRepo // 权限表数据层实例
	db           *gorm.DB                   // 全局数据库
}

// NewRolePermService 创建角色-权限关联服务层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*RolePermService - 角色-权限关联服务层指针
func NewRolePermService() *RolePermService {
	return &RolePermService{
		RolePermRepo: repository.NewRolePermRepo(),
		RoleRepo:     repository.NewRoleRepo(),
		PermRepo:     repository.NewPermissionRepo(),
		db:           db.DB,
	}
}

// CreateRolePerm 为角色分配权限（先删后加，保证权限唯一）
// 接收值：roleId - 角色ID，permIds - 权限ID列表
// 返回值：error - 错误信息
func (rp *RolePermService) CreateRolePerm(roleId int64, permIds []int64) error {
	if roleExisting, err := rp.RoleRepo.GetRoleById(roleId); err != nil || roleExisting == nil {
		return model.RoleNotExist
	}
	for _, permId := range permIds {
		if permExisting, err := rp.PermRepo.GetPermByID(permId); err != nil || permExisting == nil {
			return model.PermissionNotExist
		}
	}
	tx := rp.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	rolePermTx := rp.RolePermRepo.WithTx(tx)

	if err := rolePermTx.DeleteRolePerm(roleId); err != nil {
		tx.Rollback()
		return err
	}

	if err := rolePermTx.CreateRolePerm(roleId, permIds); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetRolePermList 根据角色ID查询关联的权限列表
// 接收值：roleId - 角色ID
// 返回值：[]*model.SysPermission - 权限列表，int64 - 权限总数，error - 错误信息
func (rp *RolePermService) GetRolePermList(roleId int64) ([]*model.SysPermission, int64, error) {
	var permList []*model.SysPermission
	permIds, err := rp.RolePermRepo.GetRolePermList(roleId)
	if err != nil {
		return nil, 0, err
	}
	for _, permId := range permIds {
		perm, err := rp.PermRepo.GetPermByID(permId)
		if err != nil {
			return nil, 0, err
		}
		permList = append(permList, perm)
	}
	return permList, int64(len(permList)), nil
}

// DeleteRolePermRel 根据角色ID删除所有关联的权限
// 接收值：roleId - 角色ID
// 返回值：error - 错误信息
func (rp *RolePermService) DeleteRolePermRel(roleId int64) error {
	if permExist, err := rp.RoleRepo.GetRoleById(roleId); err != nil || permExist == nil {
		return model.RoleNotExist
	}
	return rp.RolePermRepo.DeleteRolePerm(roleId)
}
