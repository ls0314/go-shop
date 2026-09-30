package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// UserRoleRepo 用户-角色关联表数据层实例
type UserRoleRepo struct {
	DB *gorm.DB
}

// NewUserRoleRepo 创建用户-角色关联表数据层实例
func NewUserRoleRepo(conn *gorm.DB) *UserRoleRepo {
	return &UserRoleRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (ur *UserRoleRepo) WithTx(tx *gorm.DB) *UserRoleRepo {
	return &UserRoleRepo{DB: tx}
}

// CreateUserRole 批量创建用户-角色关联,空列表不写入
func (ur *UserRoleRepo) CreateUserRole(userId int64, roleIds []int64) error {
	if len(roleIds) == 0 {
		return nil
	}
	list := make([]model.SysUserRole, 0, len(roleIds))
	for _, roleId := range roleIds {
		list = append(list, model.SysUserRole{
			UserId: userId,
			RoleId: roleId,
		})
	}
	return ur.DB.Create(&list).Error
}

// ListRoleIdsBySingleUserId 查单个用户已绑定的角色ID(与菜单树的并集查询区分)
func (ur *UserRoleRepo) ListRoleIdsBySingleUserId(userId int64) ([]int64, error) {
	var roleIds []int64
	err := ur.DB.Model(&model.SysUserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	if err != nil {
		return nil, err
	}
	return roleIds, nil
}

// DeleteByUserId 删除用户的全部角色关联
func (ur *UserRoleRepo) DeleteByUserId(userId int64) error {
	return ur.DB.Where("user_id = ?", userId).Delete(&model.SysUserRole{}).Error
}

// ListUserIdsByRoleId 查持有某角色的全部用户ID,供权限缓存失效
func (ur *UserRoleRepo) ListUserIdsByRoleId(roleId int64) ([]int64, error) {
	var userIds []int64
	err := ur.DB.Model(&model.SysUserRole{}).
		Where("role_id = ?", roleId).
		Pluck("user_id", &userIds).Error
	if err != nil {
		return nil, err
	}
	return userIds, nil
}
