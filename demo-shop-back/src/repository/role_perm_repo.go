package repository

import (
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// RolePermRepo 角色-权限关联表数据层实例
type RolePermRepo struct {
	DB *gorm.DB
}

// NewRolePermRepo 创建角色-权限关联表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*RolePermRepo - 角色-权限关联表数据层指针
func NewRolePermRepo(conn *gorm.DB) *RolePermRepo {
	return &RolePermRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*RolePermRepo - 绑定事务的角色-权限关联表数据层指针
func (rp *RolePermRepo) WithTx(tx *gorm.DB) *RolePermRepo {
	return &RolePermRepo{DB: tx}
}

// CreateRolePerm 批量创建角色-权限关联关系
// 接收值：roleId - 角色ID，permIds - 权限ID列表
// 返回值：error - 错误信息
func (rp *RolePermRepo) CreateRolePerm(roleId int64, permIds []int64) error {
	var rolePermList []model.SysRolePermission
	for _, permId := range permIds {
		rolePermList = append(rolePermList, model.SysRolePermission{
			RoleId: roleId,
			PermId: permId,
		})
	}
	if len(rolePermList) == 0 {
		return nil
	}

	return rp.DB.Create(&rolePermList).Error
}

// GetRolePermList 根据角色ID查询关联的权限ID列表
// 接收值：roleId - 角色ID
// 返回值：[]int64 - 权限ID列表，error - 错误信息
func (rp *RolePermRepo) GetRolePermList(roleId int64) ([]int64, error) {
	var permIds []int64
	err := rp.DB.Model(&model.SysRolePermission{}).
		Where("role_id = ?", roleId).
		Pluck("permission_id", &permIds).Error
	if err != nil {
		return nil, err
	}
	return permIds, nil
}

// DeleteRolePerm 根据角色ID删除所有关联的权限
// 接收值：id - 角色ID
// 返回值：error - 错误信息
func (rp *RolePermRepo) DeleteRolePerm(id int64) error {
	return rp.DB.Where("role_id = ?", id).Delete(&model.SysRolePermission{}).Error
}
