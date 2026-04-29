package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"gorm.io/gorm"
)

// RoleMenuService 角色-菜单关联服务层实例
type RoleMenuService struct {
	RoleMenuRepo *repository.RoleMenuRepo // 角色-菜单关联表数据层实例
	RoleRepo     *repository.RoleRepo     // 角色表数据层实例
	MenuRepo     *repository.MenuRepo     // 菜单表数据层实例
	db           *gorm.DB                 // 全局数据库
}

// NewRoleMenuService 创建角色-菜单关联服务层实例
// 接收值：使用全局数据库和repository初始化，故无接收值
// 返回值：*RoleMenuService - 角色-菜单关联服务层指针
func NewRoleMenuService() *RoleMenuService {
	return &RoleMenuService{
		RoleMenuRepo: repository.NewRoleMenuRepo(),
		RoleRepo:     repository.NewRoleRepo(),
		MenuRepo:     repository.NewMenuRepo(),
		db:           db.DB,
	}
}

// CreateRoleMenu 为角色分配菜单（先删后加，保证菜单唯一）
// 接收值：roleId - 角色ID，menuIds - 菜单ID列表
// 返回值：error - 错误信息
func (rm *RoleMenuService) CreateRoleMenu(roleId int64, menuIds []int64) error {
	if roleExist, err := rm.RoleRepo.GetRoleById(roleId); err != nil || roleExist == nil {
		return model.RoleNotExist
	}
	for _, menuId := range menuIds {
		if menuExist, err := rm.MenuRepo.GetMenuById(menuId); err != nil || menuExist == nil {
			return model.MenuNotExist
		}
	}
	tx := rm.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	roleMenuTx := rm.RoleMenuRepo.WithTx(tx)

	if err := roleMenuTx.DeleteRoleMenuByRoleId(roleId); err != nil {
		tx.Rollback()
		return err
	}

	if err := roleMenuTx.CreateRoleMenu(roleId, menuIds); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetRoleMenuList 根据角色ID查询关联的菜单列表
// 接收值：roleId - 角色ID
// 返回值：[]*model.SysMenu - 菜单列表，int64 - 菜单总数，error - 错误信息
func (rm *RoleMenuService) GetRoleMenuList(roleId int64) ([]*model.SysMenu, int64, error) {

	menuIds, err := rm.RoleMenuRepo.GetRoleMenuListByRoleId(roleId)
	if err != nil {
		return nil, 0, err
	}

	menuList, err := rm.MenuRepo.ListMenuByIds(menuIds)
	if err != nil {
		return nil, 0, err
	}
	return menuList, int64(len(menuList)), nil
}

// DeleteRoleMenuByRoleId 根据角色ID删除所有关联的菜单
// 接收值：roleId - 角色ID
// 返回值：error - 错误信息
func (rm *RoleMenuService) DeleteRoleMenuByRoleId(roleId int64) error {
	if roleExist, err := rm.RoleRepo.GetRoleById(roleId); err != nil || roleExist == nil {
		return model.RoleNotExist
	}
	// 开启事务
	tx := rm.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	roleMenuTxRepo := rm.RoleMenuRepo.WithTx(tx)
	// 调用数据层删除角色菜单关联
	if err := roleMenuTxRepo.DeleteRoleMenuByRoleId(roleId); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
}
