package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"gorm.io/gorm"
)

// UserDeptService 用户-部门关联服务层实例
type UserDeptService struct {
	UserDeptRepo   *repository.UserDeptRepo   // 用户-部门关联表数据层实例
	UserRepo       *repository.UserRepo       // 用户表数据层实例
	DepartmentRepo *repository.DepartmentRepo // 部门表数据层实例
	db             *gorm.DB                   // 全局数据库
}

// NewUserDeptService 创建用户-部门关联服务层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserDeptService - 用户-部门关联服务层指针
func NewUserDeptService() *UserDeptService {
	return &UserDeptService{
		UserDeptRepo:   repository.NewUserDeptRepo(),
		UserRepo:       repository.NewUserRepo(),
		DepartmentRepo: repository.NewDeptRepo(),
		db:             db.DB,
	}
}

// CreateUserDept 为用户分配部门（先删后加，设置主部门）
// 接收值：userId - 用户ID，deptIds - 部门ID列表，isPrimaryId - 主部门索引
// 返回值：error - 错误信息
func (ud *UserDeptService) CreateUserDept(userId int64, deptIds []int64, isPrimaryId int64) error {
	// 判断用户是否存在
	if userExist, err := ud.UserRepo.GetUserById(userId); err != nil || userExist == nil {
		return model.UserNotExist
	}
	// 判断部门是否存在
	for _, deptId := range deptIds {
		if deptExist, err := ud.DepartmentRepo.GetDeptById(deptId); err != nil || deptExist == nil {
			return model.DeptNotExist
		}
	}
	// 开启事务
	tx := ud.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 实例化
	userDeptTx := ud.UserDeptRepo.WithTx(tx)
	// 删除此角色关联的所有部门
	if err := userDeptTx.DeleteUserDeptByUserId(userId); err != nil {
		tx.Rollback()
		return err
	}
	// 批量创建角色部门关系
	if err := userDeptTx.CreateUserDpet(userId, deptIds, isPrimaryId); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetUserDeptList 根据用户ID查询关联的部门列表及主部门ID
// 接收值：userId - 用户ID
// 返回值：[]*model.SysDept - 部门列表，int64 - 部门总数，int64 - 主部门ID，error - 错误信息
func (ud *UserDeptService) GetUserDeptList(userId int64) ([]*model.SysDept, int64, int64, error) {
	var deptList []*model.SysDept
	// 根据用户Id获取部门Id列表
	deptIds, err := ud.UserDeptRepo.GetUserDeptListByUserId(userId)
	if err != nil {
		return nil, 0, 0, err
	}
	// 获取此用户的主部门Id
	primaryDeptId, err := ud.UserDeptRepo.GetPrimaryDeptByUserId(userId)
	if err != nil {
		return nil, 0, 0, err
	}
	// 根据部门ID列表获取其部门列表完整信息
	deptList, err = ud.DepartmentRepo.ListDeptByIds(deptIds)
	if err == nil {
		return nil, 0, 0, err
	}
	// 返回部门列表，此角色对应部门数量，主部门Id, 错误信息
	return deptList, int64(len(deptList)), primaryDeptId, nil
}

// DeleteUserDeptById 根据用户ID删除所有关联的部门
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (ud *UserDeptService) DeleteUserDeptById(userId int64) error {
	if userExist, err := ud.UserRepo.GetUserById(userId); err != nil || userExist == nil {
		return model.UserNotExist
	}
	// 开启事务
	tx := ud.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	userDeptTxRepo := ud.UserDeptRepo.WithTx(tx)
	// 调用数据层删除用户部门关联
	if err := userDeptTxRepo.DeleteUserDeptByUserId(userId); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
}
