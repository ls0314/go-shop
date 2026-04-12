package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
)

// ScopeService 数据权限范围服务层实例
type ScopeService struct {
	ScopeRope *repository.ScopeRope // 数据权限范围数据层实例
	RoleRepo  *repository.RoleRepo  // 角色数据层实例
}

// NewScopeService 创建数据权限范围服务层实例
// 接收值：
//
//	scopeRope - 数据权限范围数据层实例
//	roleRepo - 角色数据层实例
//
// 返回值：*ScopeService - 数据权限范围服务层实例指针
func NewScopeService(scopeRope *repository.ScopeRope, roleRepo *repository.RoleRepo) *ScopeService {
	return &ScopeService{
		ScopeRope: scopeRope,
		RoleRepo:  roleRepo,
	}
}

// CreateScope 创建数据权限范围
// 接收值：scope - 数据权限范围结构体指针
// 返回值：error - 错误信息
func (s *ScopeService) CreateScope(scope *model.SysScope) error {
	// 校验关联的角色是否存在
	existingRole, err := s.RoleRepo.GetRoleById(scope.RoleId)
	if err != nil {
		return err
	}
	if existingRole == nil {
		return model.ScopeNotRole
	}
	// 联合唯一校验：角色+资源类型+字段名 不能重复
	existing, err := s.ScopeRope.GetScopeByUnique(
		scope.RoleId,
		scope.ResourceType,
		scope.FieldName,
	)
	if err != nil {
		return err
	}
	if existing != nil {
		return model.ScopeExist
	}
	// 调用数据层创建数据权限范围
	return s.ScopeRope.CreateScope(scope)
}

// GetScopeById 根据ID查询数据权限范围信息
// 接收值：
//
//	id - 所查询数据权限范围唯一标识
//
// 返回值：
//
//	*model.SysScope - 数据权限范围对象指针
//	error - 错误信息
func (s *ScopeService) GetScopeById(id int64) (*model.SysScope, error) {
	// 调用数据层查询数据权限范围信息
	scope, err := s.ScopeRope.GetScopeById(id)
	if err != nil {
		return nil, model.ScopeNotExist
	}
	return scope, nil
}

// GetScopeList 分页查询数据权限范围信息（可根据资源类型查询）
// 接收值：
//
//	page - 页数,
//	pageSize - 页大小
//	resourceType - 资源类型
//
// 返回值:
//
//	[]*model.SysScope - 分页数据权限范围信息列表
//	int64 - 数据权限范围总数
//	error - 错误信息
func (s *ScopeService) GetScopeList(page, pageSize int, resourceType string) ([]*model.SysScope, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	// 调用数据层返回分页数据权限范围信息
	return s.ScopeRope.GetScopeList(page, pageSize, resourceType)

}

// UpdateScope 更新数据权限范围信息
// 接收值：
//
//	id - 更新数据权限范围唯一标识
//	scope - 更新数据权限范围部分信息
//
// 返回值：
//
//	error - 错误信息
func (s *ScopeService) UpdateScope(id int64, scope map[string]interface{}) error {
	// 查询待更新的数据权限范围是否存在
	oldScope, err := s.ScopeRope.GetScopeById(id)
	if err != nil {
		return model.ScopeNotExist
	}
	if oldScope == nil {
		return model.ScopeNotExist
	}
	// 校验关联的角色是否存在
	existRole, err := s.RoleRepo.GetRoleById(oldScope.RoleId)
	if err != nil {
		return err
	}
	if existRole == nil {
		return model.ScopeNotRole
	}

	// 用原数据权限范围信息初始化更新对象
	newScope := *oldScope
	// 配置mapstructure，用scope覆盖更新数据权限范围信息（只覆盖传入字段）
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newScope,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(scope); err != nil {
		return err
	}

	// 不允许修改数据权限范围关联的角色ID
	if oldScope.RoleId != newScope.RoleId {
		return model.ScopeIsRole
	}
	// 调用数据层更新数据权限范围信息
	return s.ScopeRope.UpdateScope(&newScope)
}

// DeleteScope 删除数据权限范围
// 接收值：id - 待删除数据权限范围唯一标识
// 返回值：error - 错误信息
func (s *ScopeService) DeleteScope(id int64) error {
	// 判断待删除的数据权限范围是否存在
	existing, err := s.ScopeRope.GetScopeById(id)
	if err != nil {
		return model.ScopeNotExist
	}
	if existing == nil {
		return model.ScopeNotExist
	}
	// 调用数据层删除数据权限范围
	return s.ScopeRope.DeleteScope(id)
}
