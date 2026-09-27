package repository

import (
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// RoleMenuRepo 角色-菜单关联表数据层实例
type RoleMenuRepo struct {
	DB *gorm.DB
}

// NewRoleMenuRepo 创建角色-菜单关联表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*RoleMenuRepo - 角色-菜单关联表数据层指针
func NewRoleMenuRepo(conn *gorm.DB) *RoleMenuRepo {
	return &RoleMenuRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*RoleMenuRepo - 绑定事务的角色-菜单关联表数据层指针
func (rm *RoleMenuRepo) WithTx(tx *gorm.DB) *RoleMenuRepo {
	return &RoleMenuRepo{DB: tx}
}

// CreateRoleMenu 批量创建角色-菜单关联关系
// 接收值：roleId - 角色ID，menuIds - 菜单ID列表
// 返回值：error - 错误信息
func (rm *RoleMenuRepo) CreateRoleMenu(roleId int64, menuIds []int64) error {
	var roleMenuList []model.SysRoleMenu
	for _, menuId := range menuIds {
		roleMenuList = append(roleMenuList, model.SysRoleMenu{
			RoleId: roleId,
			MenuId: menuId,
		})
	}
	if len(roleMenuList) == 0 {
		return nil
	}
	return rm.DB.Create(&roleMenuList).Error
}

// GetRoleMenuListByRoleIds 根据角色ID数组查询关联的菜单ID列表
// 接收值：roleIds - 角色ID数组
// 返回值：[]int64 - 菜单ID列表，error - 错误信息
func (rm *RoleMenuRepo) GetRoleMenuListByRoleIds(roleIds []int64) ([]int64, error) {
	var menuIds []int64

	if len(roleIds) == 0 {
		return menuIds, nil
	}

	err := rm.DB.Model(&model.SysRoleMenu{}).
		Where("role_id IN ?", roleIds).
		Distinct("menu_id").
		Pluck("menu_id", &menuIds).Error

	if err != nil {
		return nil, err
	}

	return menuIds, nil
}

// DeleteRoleMenuByRoleId 根据角色ID删除所有关联的菜单
// 接收值：roleId - 角色ID
// 返回值：error - 错误信息
func (rm *RoleMenuRepo) DeleteRoleMenuByRoleId(roleId int64) error {
	return rm.DB.Where("role_id = ?", roleId).Delete(&model.SysRoleMenu{}).Error
}
