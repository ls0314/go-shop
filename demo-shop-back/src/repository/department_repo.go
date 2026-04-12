package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

type DepartmentRepo struct {
	DB *gorm.DB
}

// NewDeptRepo 创建部门数据访问层实例
// 接收值：使用全局数据库，故无接收值
// 返回值: *DepartmentRepo - 部门数据访问层实例
func NewDeptRepo() *DepartmentRepo {
	return &DepartmentRepo{DB: db.DB}
}

// CreateDept 新增部门信息
// 接收值： dept - 部门实体对象
// 返回值: error - 错误信息
func (d *DepartmentRepo) CreateDept(dept *model.SysDept) error {
	return d.DB.Create(&dept).Error
}

// GetDeptById 根据ID查询部门信息
// 接收值：
//
//	id - 部门ID
//
// 返回值:
//
//	*model.SysDept - 部门信息
//	error - 错误信息
func (d *DepartmentRepo) GetDeptById(id int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.DB.First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, err
}

// GetDeptByUK 根据部门名称+父级ID查询部门（唯一索引）
// 接收值：
//
//	deptName - 部门名称
//	parentId - 父级部门ID
//
// 返回值:
//
//	*model.SysDept - 部门信息
//	error - 错误信息
func (d *DepartmentRepo) GetDeptByUK(deptName string, parentId int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.DB.Where("parent_id = ? AND dept_name = ?", parentId, deptName).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, err
}

// GetDeptList 分页查询部门信息（可根据类型查询）
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//	deptType - 部门类型
//
// 返回值:
//
//	[]model.SysDept - 分页部门信息
//	int64 - 部门总数
//	error - 错误信息
func (d *DepartmentRepo) GetDeptList(page, pageSize int, deptType string) ([]model.SysDept, int64, error) {
	var deptList []model.SysDept
	var total int64
	// 构建部门查询条件
	deptDb := d.DB.Model(&model.SysDept{}).Order("sort_order ASC, dept_id ASC")

	// 如果部门类型不为空，添加条件筛选
	if deptType != "" {
		deptDb = deptDb.Where("dept_type = ?", deptType)
	}
	// 查询部门总记录数
	if err := deptDb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 分页查询部门数据
	offset := (page - 1) * pageSize
	if err := deptDb.Limit(pageSize).Offset(offset).Find(&deptList).Error; err != nil {
		return nil, 0, err
	}
	// 返回分页部门信息、部门总数、错误信息
	return deptList, total, nil
}

// UpdateDept 更新部门信息
// 接收值： dept - 部门实体对象
// 返回值: error - 错误信息
func (d *DepartmentRepo) UpdateDept(dept *model.SysDept) error {
	return d.DB.Save(&dept).Error
}

// DeleteDept 根据ID删除部门
// 接收值： id - 部门ID
// 返回值: error - 错误信息
func (d *DepartmentRepo) DeleteDept(id int64) error {
	return d.DB.Delete(&model.SysDept{}, id).Error
}

//TODO: 检查是否存在用户与此部门关联
//func (d *DepartmentRepo) CheckDeptRelUser(deptId int64) (bool, error) {
//	var count int64
//	err := d.DB.Table("sys_dept_user").Where("dept_id=?", deptId).Count(&count).Error
//	if err != nil {
//		return false, err
//	}
//	return count > 0, nil
//}
