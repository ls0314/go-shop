package repository

import (
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// UserDeptRepo 用户-部门关联表数据层实例
type UserDeptRepo struct {
	DB *gorm.DB
}

// NewUserDeptRepo 创建用户-部门关联表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserDeptRepo - 用户-部门关联表数据层指针
func NewUserDeptRepo(conn *gorm.DB) *UserDeptRepo {
	return &UserDeptRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*UserDeptRepo - 绑定事务的用户-部门关联表数据层指针
func (ud *UserDeptRepo) WithTx(tx *gorm.DB) *UserDeptRepo {
	return &UserDeptRepo{
		DB: tx,
	}
}

// CreateUserDpet 批量创建用户-部门关联关系
// 接收值：userId - 用户ID，deptIds - 部门ID列表，isPrimaryId - 主部门索引
// 返回值：error - 错误信息
func (ud *UserDeptRepo) CreateUserDpet(userId int64, deptIds []int64, isPrimaryId int64) error {
	var userDeptList []model.SysUserDept
	for idx, deptId := range deptIds {
		isPrimary := int64(idx) == isPrimaryId
		userDeptList = append(userDeptList, model.SysUserDept{
			DeptId:    deptId,
			UserId:    userId,
			IsPrimary: isPrimary,
		})
	}
	if len(userDeptList) == 0 {
		return nil
	}
	return ud.DB.Create(&userDeptList).Error
}

// GetUserDeptListByUserId 根据用户ID查询关联的部门ID列表
// 接收值：userId - 用户ID
// 返回值：[]int64 - 部门ID列表，error - 错误信息
func (ud *UserDeptRepo) GetUserDeptListByUserId(userId int64) ([]int64, error) {
	var deptIds []int64
	err := ud.DB.Model(&model.SysUserDept{}).
		Where("user_id = ?", userId).
		Pluck("dept_id", &deptIds).Error
	if err != nil {
		return nil, err
	}
	return deptIds, err
}

// GetPrimaryDeptByUserId 根据用户ID查询主部门ID
// 接收值：userId - 用户ID
// 返回值：int64 - 主部门ID，error - 错误信息
func (ud *UserDeptRepo) GetPrimaryDeptByUserId(userId int64) (int64, error) {
	var primaryDeptId int64

	err := ud.DB.Model(&model.SysUserDept{}).
		Where("user_id = ? AND is_primary = ?", userId, true).
		Pluck("dept_id", &primaryDeptId).Error

	return primaryDeptId, err

}

// DeleteUserDeptByUserId 根据用户ID删除所有关联的部门
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (ud *UserDeptRepo) DeleteUserDeptByUserId(userId int64) error {
	return ud.DB.Where("user_id = ?", userId).Delete(&model.SysUserDept{}).Error
}
