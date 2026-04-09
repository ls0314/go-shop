package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

type RoleRepo struct {
	DB *gorm.DB
}

func NewRoleRepo() *RoleRepo {
	return &RoleRepo{DB: db.DB}
}

func (r *RoleRepo) CreateRole(role model.SysRole) error {
	return r.DB.Create(&role).Error
}

func (r *RoleRepo) GetRoleById(id int64) (*model.SysRole, error) {
	var role model.SysRole
	err := r.DB.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) GetRoleByName(name string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.DB.Where("name = ?", name).Find(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) GetRoleList(page, pageSize int, roleType string) ([]model.SysRole, int64, error) {
	var roleList []model.SysRole
	var total int64

	query := r.DB.Model(&model.SysRole{})
	if roleType != "" {
		query = query.Where("role_type=?", roleType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("role_id DESC").Limit(pageSize).Offset(offset).Find(&roleList).Error; err != nil {
		return nil, 0, err
	}
	return roleList, total, nil
}

func (r *RoleRepo) UpdateRole(role model.SysRole) error {
	return r.DB.Save(&role).Error
}

func (r *RoleRepo) DeleteRole(id int64) error {
	return r.DB.Delete(&model.SysRole{}, id).Error
}

func (m *MenuRepo) CheckRoleRelUser(roleID int64) (bool, error) {
	var count int64
	err := m.DB.Table("sys_user_role").Where("role_id=?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
