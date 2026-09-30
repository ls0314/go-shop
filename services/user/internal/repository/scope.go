package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// ScopeRepo 数据权限表数据层实例
type ScopeRepo struct {
	DB *gorm.DB
}

// NewScopeRepo 创建数据权限表数据层实例
func NewScopeRepo(conn *gorm.DB) *ScopeRepo {
	return &ScopeRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (s *ScopeRepo) WithTx(tx *gorm.DB) *ScopeRepo {
	return &ScopeRepo{DB: tx}
}

// CreateScope 创建数据权限
func (s *ScopeRepo) CreateScope(scope *model.SysScope) error {
	return s.DB.Create(scope).Error
}

// GetScopeById 根据ID查询数据权限
func (s *ScopeRepo) GetScopeById(id int64) (*model.SysScope, error) {
	var scope model.SysScope
	if err := s.DB.First(&scope, id).Error; err != nil {
		return nil, err
	}
	return &scope, nil
}

// GetScopeByUnique 按 角色ID + 资源类型 + 字段名 查唯一数据权限。
// 查不到时返回 (nil, nil) 而非 ErrRecordNotFound。
func (s *ScopeRepo) GetScopeByUnique(roleId int64, resourceType, fieldName string) (*model.SysScope, error) {
	var scope model.SysScope
	err := s.DB.Where("role_id = ? AND resource_type = ? AND field_name = ?",
		roleId, resourceType, fieldName).First(&scope).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &scope, err
}

// GetScopeList 分页查询数据权限,按 scope_id 升序
func (s *ScopeRepo) GetScopeList(page, pageSize int, resourceType string) ([]*model.SysScope, int64, error) {
	var scopes []*model.SysScope
	var total int64

	query := s.DB.Model(&model.SysScope{}).Order("scope_id ASC")
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Limit(pageSize).Offset(offset).Find(&scopes).Error; err != nil {
		return nil, 0, err
	}
	return scopes, total, nil
}

// UpdateScope 更新数据权限
func (s *ScopeRepo) UpdateScope(scope *model.SysScope) error {
	return s.DB.Save(scope).Error
}

// DeleteScope 删除数据权限
func (s *ScopeRepo) DeleteScope(id int64) error {
	return s.DB.Delete(&model.SysScope{}, id).Error
}
