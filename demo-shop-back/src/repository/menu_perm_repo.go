package repository

import (
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// MenuPermissionRepo 菜单-权限关联表数据层实例
type MenuPermissionRepo struct {
	DB *gorm.DB
}

// NewMenuPermissionRepo 创建菜单-权限关联表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*MenuPermissionRepo - 菜单-权限关联表数据层指针
func NewMenuPermissionRepo(conn *gorm.DB) *MenuPermissionRepo {
	return &MenuPermissionRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*MenuPermissionRepo - 绑定事务的菜单-权限关联表数据层指针
func (mp *MenuPermissionRepo) WithTx(tx *gorm.DB) *MenuPermissionRepo {
	return &MenuPermissionRepo{DB: tx}
}

// CreateMenuPermission 批量创建菜单-权限关联关系
// 接收值：menuId - 菜单ID，permissionIds - 权限ID列表
// 返回值：error - 错误信息
func (mp *MenuPermissionRepo) CreateMenuPermission(menuId int64, permissionIds []int64) error {
	var menuPermissionList []model.SysMenuPermission
	for _, permissionId := range permissionIds {
		menuPermissionList = append(menuPermissionList, model.SysMenuPermission{
			MenuId:       menuId,
			PermissionId: permissionId,
		})
	}
	if len(menuPermissionList) == 0 {
		return nil
	}
	return mp.DB.Create(&menuPermissionList).Error
}

// GetPermissionByMenuId 根据菜单ID查询关联的权限ID列表
// 接收值：menuId - 菜单ID
// 返回值：[]int64 - 权限ID列表，error - 错误信息
func (mp *MenuPermissionRepo) GetPermissionByMenuId(menuId int64) ([]int64, error) {
	var permissionIds []int64
	err := mp.DB.Model(&model.SysMenuPermission{}).
		Where("menu_id = ?", menuId).
		Pluck("permission_id", &permissionIds).Error
	if err != nil {
		return nil, err
	}
	return permissionIds, nil
}

// DeleteMenuPermissionByMenuId 根据菜单ID删除所有关联的权限
// 接收值：menuId - 菜单ID
// 返回值：error - 错误信息
func (mp *MenuPermissionRepo) DeleteMenuPermissionByMenuId(menuId int64) error {
	return mp.DB.Where("menu_id = ?", menuId).Delete(&model.SysMenuPermission{}).Error
}
