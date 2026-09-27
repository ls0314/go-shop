package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// MenuService 菜单表服务层实例
type MenuService struct {
	MenuRepo     *repository.MenuRepo     // 菜单表数据层实例
	RoleMenuRepo *repository.RoleMenuRepo // 角色-菜单关联表数据层实例
	UserRoleRepo *repository.UserRoleRepo // 用户-角色关联表数据层实例
	db           *gorm.DB
}

// NewMenuService 创建菜单表服务层实例
// 接收值：注入的数据库连接（由 composition root 提供）
// 返回值：*MenuService - 菜单表服务层实例指针
func NewMenuService(deps ServiceDeps) *MenuService {
	return &MenuService{
		MenuRepo:     repository.NewMenuRepo(deps.DB),
		RoleMenuRepo: repository.NewRoleMenuRepo(deps.DB),
		UserRoleRepo: repository.NewUserRoleRepo(deps.DB),
		db:           deps.DB,
	}
}

// CreateMenu 创建菜单
// 在保证同一父节点下没有相同的菜单名的情况下创建新菜单
// 接收值：menu - 菜单结构体
// 返回值：error - 错误信息
func (m *MenuService) CreateMenu(menu *model.SysMenu) error {
	//	创建联合判重 保证同一父级下没有相同的名字
	existing, _ := m.MenuRepo.GetMenuByUk(menu.ParentId, menu.MenuName)
	if existing != nil {
		return model.MenuExist
	}
	// 如果未传入Sort_id自动加到该父节点最后面
	if menu.SortOrder == 0 {
		sortId, err := m.MenuRepo.GetMaxSortId(menu.ParentId)
		if err != nil {
			return err
		}
		menu.SortOrder = sortId
	}
	// 开启事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	menuTxRepo := m.MenuRepo.WithTx(tx)
	// 调用数据层创建菜单
	if err := menuTxRepo.CreateMenu(menu); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

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
//	page - 页数
//	pageSize - 页大小
//	menuType - 菜单类型
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
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	// 调用数据层返回分页菜单信息
	return m.MenuRepo.GetMenuList(page, pageSize, menuType)
}

func (m *MenuService) MakeTree(menuList []*model.SysMenu) ([]*model.SysMenu, error) {
	// 初始化 menuMap
	menuMap := make(map[int64]*model.SysMenu)
	for _, menu := range menuList {
		menuMap[menu.MenuId] = menu
	}

	// 构建树形结构
	var treeList []*model.SysMenu
	for _, menu := range menuList {
		parent, hasParent := menuMap[menu.ParentId]
		if !hasParent {
			treeList = append(treeList, menu)
			continue
		}
		parent.Children = append(parent.Children, menu)
	}
	return treeList, nil
}

// GetMenuTreeByRoleIds 获取角色菜单树
// 接收值：roleIds - 角色ID数组
// 返回值：
//
//	[]*model.SysMenu - 相关角色下的菜单树
//	error - 错误信息
func (m *MenuService) GetMenuTreeByRoleIds(roleIds []int64) ([]*model.SysMenu, error) {
	// 获取角色关联的菜单ID列表
	menuIds, err := m.RoleMenuRepo.GetRoleMenuListByRoleIds(roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return []*model.SysMenu{}, nil
	}
	// 批量查询菜单，且已按 sort_order 排序
	menuList, err := m.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, err
	}
	treeList, err := m.MakeTree(menuList)
	if err != nil {
		return nil, err
	}
	return treeList, nil
}

// GetMenuTreeByUserId 获取用户对应的角色菜单树并集
// 接收值：userId - 角色ID
// 返回值：
//
//	[]*model.SysMenu - 相关角色下的菜单树
//	error - 错误信息
func (m *MenuService) GetMenuTreeByUserId(userId int64) ([]*model.SysMenu, error) {
	// 获取用户关联的角色ID列表
	roleIds, err := m.UserRoleRepo.GetUserRoleByUserId(userId)
	if err != nil {
		return nil, err
	}
	// 获取角色关联的菜单ID列表
	menuIds, err := m.RoleMenuRepo.GetRoleMenuListByRoleIds(roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return []*model.SysMenu{}, nil
	}
	// 批量查询菜单，且已按 sort_order 排序
	menuList, err := m.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, err
	}

	treeList, err := m.MakeTree(menuList)
	if err != nil {
		return nil, err
	}
	return treeList, nil

	//// 初始化 menuMap
	//menuMap := make(map[int64]*model.SysMenu)
	//for _, menu := range menuList {
	//	menuMap[menu.MenuId] = menu
	//}
	//
	//// 构建树形结构
	//var treeList []*model.SysMenu
	//for _, menu := range menuList {
	//	parent, hasParent := menuMap[menu.ParentId]
	//	if !hasParent {
	//		treeList = append(treeList, menu)
	//		continue
	//	}
	//	parent.Children = append(parent.Children, menu)
	//}
	//return treeList, nil
}

// UpdateMenu 更新菜单信息
// 接收值：
//
//	menuID - 更新菜单唯一标识
//	updateMenu - 更新菜单部分信息
//
// 返回值：
//
//	error - 错误信息
func (m *MenuService) UpdateMenu(menuID int64, updateMenu map[string]interface{}) error {
	// 查询所更新菜单是否存在
	olderMenu, err := m.MenuRepo.GetMenuById(menuID)
	if err != nil {
		return model.MenuNotExist
	}
	// 用原菜单信息初始化更新菜单信息
	newMenu := *olderMenu
	// 配置mapstructure，用updateMenu覆盖更新菜单信息（只覆盖传入字段）
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
	// 变更父节点自动插入队尾
	if newMenu.ParentId != olderMenu.ParentId {
		sortId, err := m.MenuRepo.GetMaxSortId(newMenu.ParentId)
		if err != nil {
			return err
		}
		newMenu.SortOrder = sortId
	}

	// 开启事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	menuTxRepo := m.MenuRepo.WithTx(tx)
	// 调用数据层更新菜单
	if err := menuTxRepo.UpdateMenu(&newMenu); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

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
	// 开启事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	menuTxRepo := m.MenuRepo.WithTx(tx)
	// 调用数据层删除菜单
	if err := menuTxRepo.DeleteMenu(id); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

}
