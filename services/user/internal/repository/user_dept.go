package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// UserDeptRepo 用户-部门关联表数据层实例
type UserDeptRepo struct {
	DB *gorm.DB
}

// NewUserDeptRepo 创建用户-部门关联表数据层实例
func NewUserDeptRepo(conn *gorm.DB) *UserDeptRepo {
	return &UserDeptRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (ud *UserDeptRepo) WithTx(tx *gorm.DB) *UserDeptRepo {
	return &UserDeptRepo{DB: tx}
}

// CreateUserDept 批量创建用户-部门关联,空列表不写入。
// primaryIndex 是 deptIds 中的下标,指向主部门;越界或为负表示无主部门。
func (ud *UserDeptRepo) CreateUserDept(userId int64, deptIds []int64, primaryIndex int64) error {
	if len(deptIds) == 0 {
		return nil
	}
	list := make([]model.SysUserDept, 0, len(deptIds))
	for idx, deptId := range deptIds {
		list = append(list, model.SysUserDept{
			UserId:    userId,
			DeptId:    deptId,
			IsPrimary: int64(idx) == primaryIndex,
		})
	}
	return ud.DB.Create(&list).Error
}

// ListDeptIdsBySingleUserId 查单个用户已绑定的部门ID(与部门树的并集查询区分)
func (ud *UserDeptRepo) ListDeptIdsBySingleUserId(userId int64) ([]int64, error) {
	var deptIds []int64
	err := ud.DB.Model(&model.SysUserDept{}).
		Where("user_id = ?", userId).
		Pluck("dept_id", &deptIds).Error
	if err != nil {
		return nil, err
	}
	return deptIds, nil
}

// GetPrimaryDeptId 查用户的主部门ID,无主部门时返回 0
func (ud *UserDeptRepo) GetPrimaryDeptId(userId int64) (int64, error) {
	var deptId int64
	err := ud.DB.Model(&model.SysUserDept{}).
		Where("user_id = ? AND is_primary = ?", userId, true).
		Limit(1).
		Pluck("dept_id", &deptId).Error
	if err != nil {
		return 0, err
	}
	return deptId, nil
}

// DeleteByUserId 删除用户的全部部门关联
func (ud *UserDeptRepo) DeleteByUserId(userId int64) error {
	return ud.DB.Where("user_id = ?", userId).Delete(&model.SysUserDept{}).Error
}
