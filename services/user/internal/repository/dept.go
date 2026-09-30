package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// DeptRepo 部门数据访问层实例。
// 部门树需要跨 sys_user_dept 取用户所属部门ID,该查询按用途并入本 repo。
type DeptRepo struct {
	DB *gorm.DB
}

// NewDeptRepo 创建部门数据访问层实例
func NewDeptRepo(conn *gorm.DB) *DeptRepo {
	return &DeptRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (d *DeptRepo) WithTx(tx *gorm.DB) *DeptRepo {
	return &DeptRepo{DB: tx}
}

// CreateDept 新增部门
func (d *DeptRepo) CreateDept(dept *model.SysDept) error {
	return d.DB.Create(dept).Error
}

// GetDeptById 根据ID查询部门
func (d *DeptRepo) GetDeptById(id int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.DB.First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// GetDeptByUK 按 父级ID + 部门名 查部门(联合唯一)
func (d *DeptRepo) GetDeptByUK(deptName string, parentId int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.DB.Where("parent_id = ? AND dept_name = ?", parentId, deptName).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// GetMaxSortId 获取此父节点下的最大排序号,无子节点时返回 0
func (d *DeptRepo) GetMaxSortId(parentId int64) (int64, error) {
	var sortId int64
	err := d.DB.Model(&model.SysDept{}).
		Where("parent_id = ?", parentId).
		Select("COALESCE(MAX(sort_order), 0)").
		Find(&sortId).Error
	if err != nil {
		return 0, err
	}
	return sortId, nil
}

// GetDeptList 分页查询部门,按 sort_order、dept_id 升序
func (d *DeptRepo) GetDeptList(page, pageSize int, deptType string) ([]model.SysDept, int64, error) {
	var deptList []model.SysDept
	var total int64

	query := d.DB.Model(&model.SysDept{}).Order("sort_order ASC, dept_id ASC")
	if deptType != "" {
		query = query.Where("dept_type = ?", deptType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&deptList).Error; err != nil {
		return nil, 0, err
	}
	return deptList, total, nil
}

// ListDeptByIds 按部门ID列表批量查询,按 sort_order 升序
func (d *DeptRepo) ListDeptByIds(deptIds []int64) ([]*model.SysDept, error) {
	var deptList []*model.SysDept
	err := d.DB.
		Where("dept_id IN ?", deptIds).
		Order("sort_order asc").
		Find(&deptList).Error

	return deptList, err
}

// UpdateDept 更新部门
func (d *DeptRepo) UpdateDept(dept *model.SysDept) error {
	return d.DB.Save(dept).Error
}

// DeleteDept 根据ID删除部门
func (d *DeptRepo) DeleteDept(id int64) error {
	return d.DB.Delete(&model.SysDept{}, id).Error
}

// CheckDeptRelUser 检查部门是否关联用户
func (d *DeptRepo) CheckDeptRelUser(deptId int64) (bool, error) {
	var count int64
	err := d.DB.Table("sys_user_dept").Where("dept_id = ?", deptId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListDeptIdsByUserId 查询用户所属的全部部门ID,供部门树取并集
func (d *DeptRepo) ListDeptIdsByUserId(userId int64) ([]int64, error) {
	var deptIds []int64
	err := d.DB.Model(&model.SysUserDept{}).
		Where("user_id = ?", userId).
		Pluck("dept_id", &deptIds).Error
	if err != nil {
		return nil, err
	}
	return deptIds, nil
}
