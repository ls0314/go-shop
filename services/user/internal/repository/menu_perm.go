package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// MenuPermRepo 菜单-权限关联表数据层实例
type MenuPermRepo struct {
	DB *gorm.DB
}

// NewMenuPermRepo 创建菜单-权限关联表数据层实例
func NewMenuPermRepo(conn *gorm.DB) *MenuPermRepo {
	return &MenuPermRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (mp *MenuPermRepo) WithTx(tx *gorm.DB) *MenuPermRepo {
	return &MenuPermRepo{DB: tx}
}

// CreateMenuPerm 批量创建菜单-权限关联,空列表不写入
func (mp *MenuPermRepo) CreateMenuPerm(menuId int64, permIds []int64) error {
	if len(permIds) == 0 {
		return nil
	}
	list := make([]model.SysMenuPermission, 0, len(permIds))
	for _, permId := range permIds {
		list = append(list, model.SysMenuPermission{
			MenuId:       menuId,
			PermissionId: permId,
		})
	}
	return mp.DB.Create(&list).Error
}

// ListPermIdsByMenuId 查菜单已绑定的权限ID
func (mp *MenuPermRepo) ListPermIdsByMenuId(menuId int64) ([]int64, error) {
	var permIds []int64
	err := mp.DB.Model(&model.SysMenuPermission{}).
		Where("menu_id = ?", menuId).
		Pluck("permission_id", &permIds).Error
	if err != nil {
		return nil, err
	}
	return permIds, nil
}

// DeleteByMenuId 删除菜单的全部权限关联
func (mp *MenuPermRepo) DeleteByMenuId(menuId int64) error {
	return mp.DB.Where("menu_id = ?", menuId).Delete(&model.SysMenuPermission{}).Error
}
