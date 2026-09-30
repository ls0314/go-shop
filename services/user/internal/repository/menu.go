package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// MenuRepo 菜单表数据层实例。
// 菜单树需要跨 sys_role_menu / sys_user_role 取关联 ID,
// 这两个查询按用途并入本 repo(而非按表拆成独立 repo)。
type MenuRepo struct {
	DB *gorm.DB
}

// NewMenuRepo 创建菜单表数据层实例
func NewMenuRepo(conn *gorm.DB) *MenuRepo {
	return &MenuRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (m *MenuRepo) WithTx(tx *gorm.DB) *MenuRepo {
	return &MenuRepo{DB: tx}
}

// CreateMenu 创建菜单
func (m *MenuRepo) CreateMenu(menu *model.SysMenu) error {
	return m.DB.Create(menu).Error
}

// GetMenuById 查询菜单信息(按菜单ID查)
func (m *MenuRepo) GetMenuById(id int64) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetMenuByUk 联合查询菜单信息(按父节点ID和菜单名查)
func (m *MenuRepo) GetMenuByUk(parentId int64, menuName string) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.Where("parent_id = ? AND menu_name = ?", parentId, menuName).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetMaxSortId 获取当前父节点下最大排序号,无子节点时返回 0
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

// GetMenuList 分页查询菜单信息(可按类型过滤)
func (m *MenuRepo) GetMenuList(page, pageSize int, menuType string) ([]model.SysMenu, int64, error) {
	var menuList []model.SysMenu
	var total int64

	query := m.DB.Model(&model.SysMenu{})
	if menuType != "" {
		query = query.Where("menu_type = ?", menuType)
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

// ListMenuByIds 按菜单ID列表批量查询,按 sort_order 升序
func (m *MenuRepo) ListMenuByIds(menuIds []int64) ([]*model.SysMenu, error) {
	var menuList []*model.SysMenu
	err := m.DB.
		Where("menu_id IN ?", menuIds).
		Order("sort_order asc").
		Find(&menuList).Error

	return menuList, err
}

// UpdateMenu 更新菜单信息
func (m *MenuRepo) UpdateMenu(menu *model.SysMenu) error {
	return m.DB.Save(menu).Error
}

// DeleteMenu 删除菜单
func (m *MenuRepo) DeleteMenu(id int64) error {
	return m.DB.Delete(&model.SysMenu{}, id).Error
}

// CheckRoleRelMenu 检查是否有角色关联该菜单
func (m *MenuRepo) CheckRoleRelMenu(menuID int64) (bool, error) {
	var count int64
	err := m.DB.Table("sys_role_menu").Where("menu_id = ?", menuID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListRoleIdsByUserId 查询用户持有的全部角色ID,供菜单树取并集
func (m *MenuRepo) ListRoleIdsByUserId(userId int64) ([]int64, error) {
	var roleIds []int64
	err := m.DB.Model(&model.SysUserRole{}).
		Where("user_id = ?", userId).
		Pluck("role_id", &roleIds).Error
	if err != nil {
		return nil, err
	}
	return roleIds, nil
}

// ListMenuIdsByRoleIds 查询这批角色关联的全部菜单ID(去重),供菜单树取并集
func (m *MenuRepo) ListMenuIdsByRoleIds(roleIds []int64) ([]int64, error) {
	var menuIds []int64
	if len(roleIds) == 0 {
		return menuIds, nil
	}
	err := m.DB.Model(&model.SysRoleMenu{}).
		Where("role_id IN ?", roleIds).
		Distinct("menu_id").
		Pluck("menu_id", &menuIds).Error
	if err != nil {
		return nil, err
	}
	return menuIds, nil
}
