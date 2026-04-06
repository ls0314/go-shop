package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// MenuRepo 菜单表数据层实例
type MenuRepo struct {
	DB *gorm.DB // 全局数据库
}

// NewMenuRepo 创建菜单表数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*MenuRepo - 菜单表数据层实例指针
func NewMenuRepo() *MenuRepo {
	return &MenuRepo{DB: db.DB}
}

// CreateMenu 创建菜单
// 接收值：menu - 菜单对象指针
// 返回值：error - 错误信息
func (m *MenuRepo) CreateMenu(menu *model.SysMenu) error {
	return m.DB.Create(menu).Error
}

// GetMenuById 查询菜单信息(按菜单ID查)
// 接收值：id - 所查询菜单唯一标识
// 返回值：
//
//	*model.SysMenu - 所查询菜单信息
//	error - 错误信息
func (m *MenuRepo) GetMenuById(id int64) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetMenuByUk 联合查询菜单信息(按父节点ID和菜单名查）
// 接收值：
//
//	parentId - 所查询菜单父节点ID
//	menuName - 所查询菜单名
//
// 返回值：
//
//	*model.SysMenu所查询菜单对象指针
//	error - 错误信息
func (m *MenuRepo) GetMenuByUk(parentId int64, menuName string) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.Where("parent_id = ? AND menu_name = ?", parentId, menuName).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetMenuList 分页查询菜单信息（可根据类型查询）
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//	menuType - 菜单类型
//
// 返回值:
//
//	[]model.SysMenu - 分页菜单信息
//	int64 - 菜单总数
//	error - 错误信息
func (m *MenuRepo) GetMenuList(page, pageSize int, menuType string) ([]model.SysMenu, int64, error) {
	var menuList []model.SysMenu
	var total int64
	// 按类型查询菜单信息
	query := m.DB.Model(&model.SysMenu{})
	if menuType != "" {
		query = query.Where("menu_type=?", menuType)
	}
	// 查总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 分页查数据
	offset := (page - 1) * pageSize
	if err := query.Order("menu_id DESC").Limit(pageSize).Offset(offset).Find(&menuList).Error; err != nil {
		return nil, 0, err
	}
	// 返回分页菜单信息, 菜单总数, 错误信息
	return menuList, total, nil
}

// GetMenuTree 获取角色菜单树
// 接收值：菜单结构体列表
// 返回值：菜单角色树
// TODO： 根据角色ID构建菜单角色树
//func (m *MenuRepo) GetMenuTree() ([]model.SysMenu, error) {
//	var menuList []model.SysMenu
//
//	err := m.DB.Order("sort_order asc").Find(&menuList).Error
//	if err != nil {
//		return nil, err
//	}
//
//	menuMap := make(map[int64]model.SysMenu)
//	for _, it := range menuList {
//		menuMap[it.MenuId] = it
//	}
//
//	var treeList []model.SysMenu
//	for i := range menuList {
//		item := &menuList[i]
//
//		parent, hasParent := menuMap[item.ParentId]
//		if !hasParent {
//			treeList = append(treeList, *item)
//			continue
//		}
//
//		parent.Children = append(parent.Children, *item)
//		menuMap[parent.MenuId] = parent
//	}
//	return treeList, nil
//}

// UpdateMenu 更新菜单信息
// 接收值：menu - 菜单对象指针
// 返回值：error - 错误信息
func (m *MenuRepo) UpdateMenu(menu *model.SysMenu) error {
	return m.DB.Save(menu).Error
}

// DeleteMenu 删除菜单
// 接收值：id - 待删除菜单唯一标识
// 返回值：error - 错误信息
func (m *MenuRepo) DeleteMenu(id int64) error {
	return m.DB.Delete(&model.SysMenu{}, id).Error
}

// CheckRoleRelMenu 检查是否有角色关联该菜单
// 接收值：
//
//	menuID - 所查询菜单唯一标识
//
// 返回值：
//
//	bool - 该菜单是否有关联角色
//	error - 错误信息
//
// TODO： 检查查是否有角色关联该菜单
func (m *MenuRepo) CheckRoleRelMenu(menuID int64) (bool, error) {
	var count int64
	err := m.DB.Table("sys_role_menu").Where("menu_id=?", menuID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
