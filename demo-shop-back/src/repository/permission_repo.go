package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

type PermissionRepo struct {
	DB *gorm.DB
}

func NewPermissionRepo() *PermissionRepo {
	return &PermissionRepo{DB: db.DB}
}

// CreatePerm 创建权限
func (r *PermissionRepo) CreatePerm(perm *model.SysPermission) error {
	return r.DB.Create(perm).Error
}

// GetPermByID 按权限ID查
func (r *PermissionRepo) GetPermByID(id int64) (*model.SysPermission, error) {
	var perm model.SysPermission
	err := r.DB.First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GetPermByCode 按权限代码查
func (r *PermissionRepo) GetPermByCode(name string) (*model.SysPermission, error) {
	var perm model.SysPermission
	err := r.DB.Where("permission_code = ?", name).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GetPermList 分页查询
func (r *PermissionRepo) GetPermList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	var perms []model.SysPermission
	var total int64

	query := r.DB.Model(&model.SysPermission{})
	if permType != "" {
		query = query.Where("permission_type = ?", permType)
	}

	// 查总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查数据
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&perms).Error; err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

// UpdataPerm 更新权限
func (r *PermissionRepo) UpdataPerm(perm *model.SysPermission) error {
	return r.DB.Save(perm).Error
}

// DeletePerm 删除权限
func (r *PermissionRepo) DeletePerm(id int64) error {
	return r.DB.Delete(&model.SysPermission{}, id).Error
}

// CheckRoleRelPerm 删除前检查是否有角色关联该权限
func (r *PermissionRepo) CheckRoleRelPerm(permID int64) (bool, error) {
	var count int64
	err := r.DB.Table("sys_role_permission").Where("permission_id = ?", permID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
