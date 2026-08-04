package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// UserRoleRepo 用户-角色关联表数据层实例
type UserRoleRepo struct {
	DB *gorm.DB
}

// NewUserRoleRepo 创建用户-角色关联表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserRoleRepo - 用户-角色关联表数据层指针
func NewUserRoleRepo() *UserRoleRepo {
	return &UserRoleRepo{DB: db.DB}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*UserRoleRepo - 绑定事务的用户-角色关联表数据层指针
func (ur *UserRoleRepo) WithTx(tx *gorm.DB) *UserRoleRepo {
	return &UserRoleRepo{DB: tx}
}

// CreateUserRole 批量创建用户-角色关联关系
// 接收值：userId - 用户ID，roleIds - 角色ID列表
// 返回值：error - 错误信息
func (ur *UserRoleRepo) CreateUserRole(userId int64, roleIds []int64) error {
	var userRoleList []model.SysUserRole
	for _, roleId := range roleIds {
		userRoleList = append(userRoleList, model.SysUserRole{
			UserId: userId,
			RoleId: roleId,
		})
	}
	if len(userRoleList) == 0 {
		return nil
	}
	return ur.DB.Create(&userRoleList).Error
}

// GetUserRoleByUserId 根据用户ID查询关联的角色ID列表
// 接收值：userId - 用户ID
// 返回值：[]int64 - 角色ID列表，error - 错误信息
func (ur *UserRoleRepo) GetUserRoleByUserId(userId int64) ([]int64, error) {
	var roleIds []int64
	err := ur.DB.Model(&model.SysUserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	if err != nil {
		return nil, err
	}
	return roleIds, nil
}

// DeleteUserRoleByUserId 根据用户ID删除所有关联的角色
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (ur *UserRoleRepo) DeleteUserRoleByUserId(userId int64) error {
	return ur.DB.Where("user_id = ?", userId).Delete(&model.SysUserRole{}).Error
}

// GetUserIdsByRoleId 查询持有某角色的全部用户ID
func (ur *UserRoleRepo) GetUserIdsByRoleId(roleId int64) ([]int64, error) {
	var userIds []int64
	err := ur.DB.Table("sys_user_role").Where("role_id = ?", roleId).Pluck("user_id", &userIds).Error
	return userIds, err
}
