package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
)

type MenuService struct {
	MenuRepo *repository.MenuRepo
}

func NewMenuService(menuRepo *repository.MenuRepo) *MenuService {
	return &MenuService{MenuRepo: menuRepo}
}

func (m *MenuService) CreateMenu(menu *model.SysMenu) error {
	//	创建联合判重 保证同一父级下没有相同的名字
	existing, _ := m.MenuRepo.GetMenuByUk(menu.ParentId, menu.MenuName)
	if existing != nil {
		return model.MenuExist
	}
	return m.MenuRepo.CreateMenu(menu)
}

func (m *MenuService) GetMenu(id int64) (*model.SysMenu, error) {
	menu, err := m.MenuRepo.GetMenuById(id)
	if err != nil {
		return nil, model.MenuNotExist
	}
	return menu, nil
}

func (m *MenuService) GetMenuList(page, pageSize int, permType string) ([]model.SysMenu, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return m.MenuRepo.GetMenuList(page, pageSize, permType)
}

func (m *MenuService) UpdateMenu(menu *model.SysMenu) error {

	olderMenu, err := m.MenuRepo.GetMenuById(menu.ParentId)
	if err != nil {
		return model.MenuNotExist
	}

	if menu.MenuName == "" {
		menu.MenuName = olderMenu.MenuName
	}

	if menu.MenuName != olderMenu.MenuName {
		_, err = m.MenuRepo.GetMenuByUk(menu.ParentId, menu.MenuName)
		if err != nil {
			return model.MenuExist
		}
	}

	return m.MenuRepo.UpdateMenu(menu)
}

func (m *MenuService) DeleteMenu(id int64) error {
	existing, _ := m.MenuRepo.GetMenuById(id)
	if existing == nil {
		return model.MenuNotExist
	}
	hasRel, err := m.MenuRepo.CheckRoleRelMenu(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.MenuHasRel
	}

	return m.MenuRepo.DeleteMenu(id)
}
