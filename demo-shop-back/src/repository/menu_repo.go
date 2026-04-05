package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

type MenuRepo struct {
	DB *gorm.DB
}

func NewMenuRepo() *MenuRepo {
	return &MenuRepo{DB: db.DB}
}

func (m *MenuRepo) CreateMenu(menu *model.SysMenu) error {
	return m.DB.Create(menu).Error
}

func (m *MenuRepo) GetMenuById(id int64) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (m *MenuRepo) GetMenuByUk(parentid int64, menuname string) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.Where("parent_id = ? AND menu_name = ?", parentid, menuname).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (m *MenuRepo) GetMenuList(page, pageSize int, menuType string) ([]model.SysMenu, int64, error) {
	var menuList []model.SysMenu
	var total int64

	query := m.DB.Model(&model.SysMenu{})
	if menuType != "" {
		query = query.Where("menu_type=?", menuType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("menu_id DESC").Limit(pageSize).Offset(offset).Find(&menuList).Error; err != nil {
		return nil, 0, err
	}
	return menuList, total, nil
}

func (m *MenuRepo) GetMenuTree() ([]model.SysMenu, error) { // 可以优化 完成RBAC后思考代码优化过程
	var menuList []model.SysMenu

	err := m.DB.Order("sort_order asc").Find(&menuList).Error
	if err != nil {
		return nil, err
	}

	menuMap := make(map[int64]model.SysMenu)
	for _, it := range menuList {
		menuMap[it.MenuId] = it
	}

	var treeList []model.SysMenu
	for i := range menuList {
		item := &menuList[i]

		parent, hasParent := menuMap[item.ParentId]
		if !hasParent {
			treeList = append(treeList, *item)
			continue
		}

		parent.Children = append(parent.Children, *item)
		menuMap[parent.MenuId] = parent
	}
	return treeList, nil
}

func (m *MenuRepo) UpdateMenu(menu *model.SysMenu) error {
	return m.DB.Save(menu).Error
}

func (m *MenuRepo) DeleteMenu(id int64) error {
	return m.DB.Delete(&model.SysMenu{}, id).Error
}

func (m *MenuRepo) CheckRoleRelMenu(menuID int64) (bool, error) {
	var count int64
	err := m.DB.Table("sys_role_menu").Where("menu_id=?", menuID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
