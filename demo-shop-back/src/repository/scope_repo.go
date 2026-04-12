package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// ScopeRope 数据权限表数据层实例
type ScopeRope struct {
	DB *gorm.DB // 数据库连接实例
}

// NewScopeRepo 创建数据权限表数据层实例
// 返回值：*SysScope - 数据权限表数据层指针
func NewScopeRepo() *ScopeRope {
	return &ScopeRope{DB: db.DB}
}

// CreateScope 创建数据权限
// 接收值：scope - 数据权限结构体指针
// 返回值：error - 错误信息
func (s *ScopeRope) CreateScope(scope *model.SysScope) error {
	return s.DB.Create(scope).Error
}

// GetScopeById 根据ID查询数据权限信息
// 接收值：id - 数据权限唯一标识
// 返回值：*model.SysScope - 数据权限对象指针，error - 错误信息
func (s *ScopeRope) GetScopeById(id int64) (*model.SysScope, error) {
	var scope model.SysScope
	// 根据ID查询单条数据权限记录
	if err := s.DB.First(&scope, id).Error; err != nil {
		return nil, err
	}
	return &scope, nil
}

// GetScopeByUnique 根据 角色ID+资源类型+字段名 查询唯一数据权限
func (s *ScopeRope) GetScopeByUnique(roleId int64, resourceType, fieldName string) (*model.SysScope, error) {
	var scope model.SysScope
	err := s.DB.Where("role_id = ? AND resource_type = ? AND field_name = ?",
		roleId, resourceType, fieldName).First(&scope).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &scope, err
}

// GetScopeList 分页查询数据权限列表
// 支持根据资源类型筛选
// 接收值：page - 页数，pageSize - 每页条数，resourceType - 资源类型
// 返回值：[]*model.SysScope - 数据权限列表，int64 - 总条数，error - 错误信息
func (s *ScopeRope) GetScopeList(page, pageSize int, resourceType string) ([]*model.SysScope, int64, error) {
	var scopes []*model.SysScope
	var total int64

	// 初始化查询，按权限ID升序排序
	scopesDb := s.DB.Model(&model.SysScope{}).Order("scope_id ASC")

	// 资源类型不为空时添加筛选条件
	if resourceType != "" {
		scopesDb.Where("resource_type = ?", resourceType)
	}

	// 查询总记录数
	if err := scopesDb.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算分页偏移量
	offset := (page - 1) * pageSize

	// 分页查询数据权限列表
	if err := scopesDb.Limit(pageSize).Offset(offset).Find(&scopes).Error; err != nil {
		return nil, 0, err
	}
	return scopes, total, nil
}

// UpdateScope 更新数据权限信息
// 接收值：scope - 待更新的数据权限结构体指针
// 返回值：error - 错误信息
func (s *ScopeRope) UpdateScope(scope *model.SysScope) error {
	return s.DB.Save(scope).Error
}

// DeleteScope 删除数据权限
// 接收值：id - 待删除的数据权限唯一标识
// 返回值：error - 错误信息
func (s *ScopeRope) DeleteScope(id int64) error {
	return s.DB.Delete(&model.SysScope{}, id).Error
}
