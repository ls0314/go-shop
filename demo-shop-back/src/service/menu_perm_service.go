package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"gorm.io/gorm"
)

// MenuPermissionService 菜单-权限关联服务层实例
type MenuPermissionService struct {
	MenuPermissionRepo *repository.MenuPermissionRepo
	MenuRepo           *repository.MenuRepo
	PermissionRepo     *repository.PermissionRepo
	db                 *gorm.DB // 全局数据库
}

// NewMenuPermissionService 创建菜单-权限关联服务层实例
// 接收值：conn - 数据库连接（由调用方注入）
// 返回值：*MenuPermissionService - 菜单-权限关联服务层指针
func NewMenuPermissionService() *MenuPermissionService {
	return &MenuPermissionService{
		MenuPermissionRepo: repository.NewMenuPermissionRepo(db.DB),
		MenuRepo:           repository.NewMenuRepo(db.DB),
		PermissionRepo:     repository.NewPermissionRepo(db.DB),
		db:                 db.DB,
	}
}

// CreateMenuPermission 为菜单分配权限（先删后加，保证权限唯一）
// 接收值：menuId - 菜单ID，permissionIds - 权限ID列表
// 返回值：error - 错误信息
func (mp *MenuPermissionService) CreateMenuPermission(menuId int64, permissionIds []int64) error {
	// 校验菜单是否存在
	if menuExisting, err := mp.MenuRepo.GetMenuById(menuId); err != nil || menuExisting == nil {
		return model.MenuNotExist
	}

	// 校验所有权限是否存在
	for _, permissionId := range permissionIds {
		if permissionExisting, err := mp.PermissionRepo.GetPermByID(permissionId); err != nil || permissionExisting == nil {
			return model.PermissionNotExist
		}
	}

	// 开启事务
	tx := mp.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	menuPermissionTx := mp.MenuPermissionRepo.WithTx(tx)

	// 先删除该菜单原有所有权限关联
	if err := menuPermissionTx.DeleteMenuPermissionByMenuId(menuId); err != nil {
		tx.Rollback()
		return err
	}

	// 批量创建新的菜单权限关联
	if err := menuPermissionTx.CreateMenuPermission(menuId, permissionIds); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetMenuPermissionList 根据菜单ID查询关联的权限列表
// 接收值：menuId - 菜单ID
// 返回值：[]*model.SysPermission - 权限列表，int64 - 权限总数，error - 错误信息
func (mp *MenuPermissionService) GetMenuPermissionList(menuId int64) ([]*model.SysPermission, int64, error) {
	var permissionList []*model.SysPermission
	permissionIds, err := mp.MenuPermissionRepo.GetPermissionByMenuId(menuId)
	if err != nil {
		return nil, 0, err
	}

	// 根据权限ID列表查询权限详情
	for _, permissionId := range permissionIds {
		permission, err := mp.PermissionRepo.GetPermByID(permissionId)
		if err != nil {
			return nil, 0, err
		}
		permissionList = append(permissionList, permission)
	}
	return permissionList, int64(len(permissionList)), nil
}

// DeleteMenuPermission 根据菜单ID删除所有关联的权限
// 接收值：menuId - 菜单ID
// 返回值：error - 错误信息
func (mp *MenuPermissionService) DeleteMenuPermission(menuId int64) error {
	// 校验菜单是否存在
	if exist, err := mp.MenuRepo.GetMenuById(menuId); err != nil || exist == nil {
		return model.RelNotExist
	}
	// 开启事务
	tx := mp.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	menuPermTxRepo := mp.MenuPermissionRepo.WithTx(tx)
	// 调用数据层删除菜单权限关联
	if err := menuPermTxRepo.DeleteMenuPermissionByMenuId(menuId); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

}
