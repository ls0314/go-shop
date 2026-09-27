package service

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"fmt"

	"gorm.io/gorm"
)

// UserRoleService 用户-角色关联服务层实例
type UserRoleService struct {
	UserRoleRepo *repository.UserRoleRepo // 用户-角色关联表数据层实例
	UserRepo     *repository.UserRepo     // 用户表数据层实例
	RoleRepo     *repository.RoleRepo     // 角色表数据层实例
	db           *gorm.DB                 // 全局数据库
}

// NewUserRoleService 创建用户-角色关联服务层实例
// 接收值：conn - 数据库连接（由调用方注入）
// 返回值：*UserRoleService - 用户-角色关联服务层指针
func NewUserRoleService() *UserRoleService {
	return &UserRoleService{
		UserRoleRepo: repository.NewUserRoleRepo(db.DB),
		UserRepo:     repository.NewUserRepo(db.DB),
		RoleRepo:     repository.NewRoleRepo(db.DB),
		db:           db.DB,
	}
}

// CreateUserRole 为用户分配角色（先删后加，保证角色唯一）
// 接收值：userId - 用户ID，roleIds - 角色ID列表
// 返回值：error - 错误信息
func (ur *UserRoleService) CreateUserRole(userId int64, roleIds []int64) error {
	// 判断用户是否存在
	if userExisting, err := ur.UserRepo.GetUserById(userId); err != nil || userExisting == nil {
		return model.UserNotExist
	}

	// 判断角色是否存在
	for _, roleId := range roleIds {
		if roleExisting, err := ur.RoleRepo.GetRoleById(roleId); err != nil || roleExisting == nil {
			return model.RoleNotExist
		}
	}

	// 开启事务
	tx := ur.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 实例化
	userRoleTx := ur.UserRoleRepo.WithTx(tx)

	// 删除此用户关联的所有角色
	if err := userRoleTx.DeleteUserRoleByUserId(userId); err != nil {
		tx.Rollback()
		return err
	}

	// 批量创建用户角色关系
	if err := userRoleTx.CreateUserRole(userId, roleIds); err != nil {
		tx.Rollback()
		return err
	}

	err := tx.Commit().Error
	if err != nil {
		return err
	}
	// 用户角色变更后失效该用户的权限缓存
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("user:perm:%d", userId))
	}
	return nil

}

// GetUserRoleList 根据用户ID查询关联的角色列表
// 接收值：userId - 用户ID
// 返回值：[]*model.SysRole - 角色列表，int64 - 角色总数，error - 错误信息
func (ur *UserRoleService) GetUserRoleList(userId int64) ([]*model.SysRole, int64, error) {
	var roleList []*model.SysRole
	// 根据用户Id获取角色Id列表
	roleIds, err := ur.UserRoleRepo.GetUserRoleByUserId(userId)
	if err != nil {
		return nil, 0, err
	}

	// 根据角色ID列表获取其角色列表完整信息
	for _, roleId := range roleIds {
		role, err := ur.RoleRepo.GetRoleById(roleId)
		if err != nil {
			return nil, 0, err
		}
		roleList = append(roleList, role)
	}
	return roleList, int64(len(roleList)), nil
}

// DeleteUserRole 根据用户ID删除所有关联的角色
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (ur *UserRoleService) DeleteUserRole(userId int64) error {
	// 判断用户是否存在
	if exist, err := ur.UserRepo.GetUserById(userId); err != nil || exist == nil {
		return model.RelNotExist
	}
	// 开启事务
	tx := ur.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	userRoleTxRepo := ur.UserRoleRepo.WithTx(tx)
	// 调用数据层删除用户角色关联
	if err := userRoleTxRepo.DeleteUserRoleByUserId(userId); err != nil {
		tx.Rollback()
		return err
	}
	err := tx.Commit().Error
	if err != nil {
		return err
	}
	// 用户角色变更后失效该用户的权限缓存
	if cache := infra.GetCache(); cache != nil {
		_ = cache.Del(context.Background(), fmt.Sprintf("user:perm:%d", userId))
	}
	return nil
}
