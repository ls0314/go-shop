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

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*MenuRepo - 绑定事务的菜单表数据层指针
func (m *MenuRepo) WithTx(tx *gorm.DB) *MenuRepo {
	return &MenuRepo{DB: tx}
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

// GetMaxSortId 获取当前父节点下最大排序号
// 接收值 : parentId - 所查询的父节点Id
// 返回值 :
//
//	int64 - 查询到的该父节点下的最大排序号
//	error - 错误信息
func (m *MenuRepo) GetMaxSortId(parentId int64) (int64, error) {
	var sortId int64
	err := m.DB.Model(&model.SysMenu{}).
		Where("parent_id = ?", parentId).
		Select("COALESCE(MAX(sort_order), 0)").
		Find(&sortId).Error
	if err != nil {
		return 0, err
	}
	return sortId, nil
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

// ListMenuByIds 根据菜单ID列表批量查询菜单，并按 sort_order 排序
// 接收值：menuIds - 菜单ID列表
// 返回值：[]*model.SysMenu - 菜单列表，error - 错误信息
func (m *MenuRepo) ListMenuByIds(menuIds []int64) ([]*model.SysMenu, error) {
	var menuList []*model.SysMenu
	err := m.DB.
		Where("menu_id IN ?", menuIds).
		Order("sort_order asc").
		Find(&menuList).Error

	return menuList, err
}

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
func (m *MenuRepo) CheckRoleRelMenu(menuID int64) (bool, error) {
	var count int64
	err := m.DB.Table("sys_role_menu").Where("menu_id=?", menuID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
