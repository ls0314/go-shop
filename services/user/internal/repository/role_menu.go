package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// RoleMenuRepo 角色-菜单关联表数据层实例
type RoleMenuRepo struct {
	DB *gorm.DB
}

// NewRoleMenuRepo 创建角色-菜单关联表数据层实例
func NewRoleMenuRepo(conn *gorm.DB) *RoleMenuRepo {
	return &RoleMenuRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (rm *RoleMenuRepo) WithTx(tx *gorm.DB) *RoleMenuRepo {
	return &RoleMenuRepo{DB: tx}
}

// CreateRoleMenu 批量创建角色-菜单关联,空列表不写入
func (rm *RoleMenuRepo) CreateRoleMenu(roleId int64, menuIds []int64) error {
	if len(menuIds) == 0 {
		return nil
	}
	list := make([]model.SysRoleMenu, 0, len(menuIds))
	for _, menuId := range menuIds {
		list = append(list, model.SysRoleMenu{
			RoleId: roleId,
			MenuId: menuId,
		})
	}
	return rm.DB.Create(&list).Error
}

// ListMenuIdsBySingleRoleId 查单个角色已绑定的菜单ID(与菜单树的并集查询区分)
func (rm *RoleMenuRepo) ListMenuIdsBySingleRoleId(roleId int64) ([]int64, error) {
	var menuIds []int64
	err := rm.DB.Model(&model.SysRoleMenu{}).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &menuIds).Error
	if err != nil {
		return nil, err
	}
	return menuIds, nil
}

// DeleteByRoleId 删除角色的全部菜单关联
func (rm *RoleMenuRepo) DeleteByRoleId(roleId int64) error {
	return rm.DB.Where("role_id = ?", roleId).Delete(&model.SysRoleMenu{}).Error
}
