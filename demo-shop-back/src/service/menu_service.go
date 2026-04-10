package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
)

// MenuService 菜单表服务层实例
type MenuService struct {
	MenuRepo *repository.MenuRepo // 菜单表服务层实例
}

// NewMenuService 创建菜单表服务层实例
// 接收值： menuRepo - 权限表数据层实例
// 返回值： *MenuService - 菜单表服务层实例指针
func NewMenuService(menuRepo *repository.MenuRepo) *MenuService {
	return &MenuService{MenuRepo: menuRepo}
}

// CreateMenu 创建菜单
// 在保证同一父节点下没有相同的菜单名的情况下创建新菜单
// 接收值： menu - 菜单结构体
// 返回值： error - 错误信息
func (m *MenuService) CreateMenu(menu *model.SysMenu) error {
	//	创建联合判重 保证同一父级下没有相同的名字
	existing, _ := m.MenuRepo.GetMenuByUk(menu.ParentId, menu.MenuName)
	if existing != nil {
		return model.MenuExist
	}
	//  调用数据层创建菜单
	return m.MenuRepo.CreateMenu(menu)
}

// GetMenu 查询菜单信息（根据菜单ID）
// 接收值：
//
//	id - 所查询菜单唯一标识
//
// 返回值：
//
//	*model.SysMenu - 菜单对象指针
//	error - 错误信息
func (m *MenuService) GetMenu(id int64) (*model.SysMenu, error) {
	//调用数据层查询菜单信息
	menu, err := m.MenuRepo.GetMenuById(id)
	if err != nil {
		return nil, model.MenuNotExist
	}
	return menu, nil
}

// GetMenuList 分页查询菜单信息（可根据菜单类型查询）
// 接收值：
//
//	page - 页数,
//	pageSize - 页大小
//	permType - 菜单类型
//
// 返回值:
//
//	[]model.SysMenu - 分页菜单信息列表
//	int64 - 菜单总数
//	error - 错误信息
func (m *MenuService) GetMenuList(page, pageSize int, menuType string) ([]model.SysMenu, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	// 调用数据层返回分页菜单信息
	return m.MenuRepo.GetMenuList(page, pageSize, menuType)
}

// GetMenuTree 获取角色菜单树
// TODO： 通过角色ID获取其对应的菜单树
//func (m *MenuService) GetMenuTree() (*[]model.SysMenu, error) {
//	menuTree, err := m.MenuRepo.GetMenuTree()
//	if err != nil {
//		return nil, model.MenuNotExist
//	}
//	return &menuTree, nil
//}

// UpdateMenu 更新权菜单信息
// 接收值：
//
//	menuID - 更新菜单唯一标识
//	updateMenu - 更新菜单部分信息
//
// 返回值：
//
//	error - 错误信息
func (m *MenuService) UpdateMenu(menuID int64, updateMenu map[string]interface{}) error {
	// 查询所更新权限是否存在
	olderMenu, err := m.MenuRepo.GetMenuById(menuID)
	if err != nil {
		return model.MenuNotExist
	}
	// 用原菜单信息初始化更新菜单信息
	newMenu := *olderMenu
	// 配置mapstructure，用updateMenu覆盖更新权限信息（只覆盖传入字段）
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newMenu,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(updateMenu); err != nil {
		return err
	}

	//联合唯一索引校验（同父节点下菜单名不能重复）
	if newMenu.MenuName != olderMenu.MenuName {
		existing, _ := m.MenuRepo.GetMenuByUk(newMenu.ParentId, newMenu.MenuName)
		if existing != nil {
			return model.MenuExist
		}
	}
	// 调用数据层更新菜单部分信息
	return m.MenuRepo.UpdateMenu(&newMenu)
}

// DeleteMenu 删除菜单
// 接收值：id - 待删除菜单唯一标识
// 返回值：error - 错误信息
func (m *MenuService) DeleteMenu(id int64) error {
	// 判断待删除菜单是否存在
	existing, _ := m.MenuRepo.GetMenuById(id)
	if existing == nil {
		return model.MenuNotExist
	}
	// 判断菜单是否还存在角色关联
	hasRel, err := m.MenuRepo.CheckRoleRelMenu(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.MenuHasRel
	}
	// 调用数据层删除菜单
	return m.MenuRepo.DeleteMenu(id)
}
