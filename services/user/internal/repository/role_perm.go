package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// RolePermRepo 角色-权限关联表数据层实例
type RolePermRepo struct {
	DB *gorm.DB
}

// NewRolePermRepo 创建角色-权限关联表数据层实例
func NewRolePermRepo(conn *gorm.DB) *RolePermRepo {
	return &RolePermRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (rp *RolePermRepo) WithTx(tx *gorm.DB) *RolePermRepo {
	return &RolePermRepo{DB: tx}
}

// CreateRolePerm 批量创建角色-权限关联,空列表不写入
func (rp *RolePermRepo) CreateRolePerm(roleId int64, permIds []int64) error {
	if len(permIds) == 0 {
		return nil
	}
	list := make([]model.SysRolePermission, 0, len(permIds))
	for _, permId := range permIds {
		list = append(list, model.SysRolePermission{
			RoleId: roleId,
			PermId: permId,
		})
	}
	return rp.DB.Create(&list).Error
}

// ListPermIdsByRoleId 查角色已绑定的权限ID
func (rp *RolePermRepo) ListPermIdsByRoleId(roleId int64) ([]int64, error) {
	var permIds []int64
	err := rp.DB.Model(&model.SysRolePermission{}).
		Where("role_id = ?", roleId).
		Pluck("permission_id", &permIds).Error
	if err != nil {
		return nil, err
	}
	return permIds, nil
}

// DeleteByRoleId 删除角色的全部权限关联
func (rp *RolePermRepo) DeleteByRoleId(roleId int64) error {
	return rp.DB.Where("role_id = ?", roleId).Delete(&model.SysRolePermission{}).Error
}
