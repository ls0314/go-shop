package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
)

// RoleService 角色对象服务层实例
type RoleService struct {
	RoleRepo *repository.RoleRepo // 数据层角色对象指针
}

// NewRoleService 新建服务层角色对象实例
// 接收值：roleRepo - 数据层角色对象指针
// 返回值：*RoleService - 服务层角色对象指针
func NewRoleService(roleRepo *repository.RoleRepo) *RoleService {
	return &RoleService{roleRepo}
}

// CreateRole 创建角色
// 接收值：role - 角色对象指针
// 返回值：error - 错误信息
func (r *RoleService) CreateRole(role *model.SysRole) error {
	// 保证新创建对象的角色名未被使用
	existing, _ := r.RoleRepo.GetRoleByName(role.RoleName)
	if existing != nil {
		return model.RoleExist
	}
	// 调用数据层创建角色对象
	err := r.RoleRepo.CreateRole(role)
	if err != nil {
		return err
	}
	return nil
}

// GetRoleById 查询角色信息(按ID查询)
// 接收值：
//
//	id - 待查询角色唯一标识
//
// 返回值：
//
//	*model.SysRole - 被查询角色对象指针
//	error - 错误信息
func (r *RoleService) GetRoleById(id int64) (*model.SysRole, error) {
	role, err := r.RoleRepo.GetRoleById(id)
	if err != nil {
		return nil, model.RoleNotExist
	}
	return role, nil
}

// GetRoleByName 查询角色信息(按角色名查)
// 接收值：
//
//	name - 待查询角色名
//
// 返回值：
//
//	*model.SysRole - 被查询角色对象指针
//	error - 错误信息
func (r *RoleService) GetRoleByName(name string) (*model.SysRole, error) {
	role, err := r.RoleRepo.GetRoleByName(name)
	if err != nil {
		return nil, model.RoleNotExist
	}
	return role, nil
}

// GetRoleList 分页查询角色信息（可根据角色类型查询）
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//	roleType - 角色类型
//
// 返回值:
//
//	[]model.SysRole - 分页角色信息列表
//	int64 - 角色总数
//	error - 错误信息
func (r *RoleService) GetRoleList(page, pageSize int, roleType string) ([]model.SysRole, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	// 调用数据层返回分页角色信息
	return r.RoleRepo.GetRoleList(page, pageSize, roleType)
}

// UpdateRole 更新角色信息
// 接收值：
//
//	id - 更新角色唯一标识
//	updateRole - 待更新角色部分信息
//
// 返回值：
//
//	error - 错误信息
func (r *RoleService) UpdateRole(id int64, updateRole map[string]interface{}) error {
	// 查询所更新角色是否存在
	oldRole, err := r.RoleRepo.GetRoleById(id)
	if err != nil {
		return model.RoleNotExist
	}
	// 禁止修改系统角色信息
	if oldRole.IsSystem {
		return model.RoleIsSystem
	}
	// 用原角色信息初始化更新角色信息
	newRole := *oldRole
	// 配置mapstructure，用updateRole覆盖更新角色信息（只覆盖传入字段）
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newRole,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(updateRole); err != nil {
		return err
	}
	// 保证更新的角色名未被使用
	if newRole.RoleName != oldRole.RoleName {
		existing, _ := r.RoleRepo.GetRoleByName(newRole.RoleName)
		if existing != nil {
			return model.RoleExist
		}
	}
	// 调用数据层更新角色部分信息
	return r.RoleRepo.UpdateRole(&newRole)
}

// DeleteRole 删除角色
// 接收值：id - 待删除角色唯一标识
// 返回值：error - 错误信息
func (r *RoleService) DeleteRole(id int64) error {
	// 检查待删除角色是否存在
	existing, _ := r.RoleRepo.GetRoleById(id)
	if existing == nil {
		return model.RoleNotExist
	}
	// 判断待删除角色是否是系统角色
	if existing.IsSystem {
		return model.RoleIsSystem
	}
	// 检查角色是否存在用户关联
	if userHasRel, err := r.RoleRepo.CheckRoleRelUser(id); err != nil || userHasRel {
		return model.UserHasRel
	}
	// 检查角色是否存在菜单关联
	if menuHasRel, err := r.RoleRepo.CheckRoleRelMenu(id); err != nil || menuHasRel {
		return model.MenuHasRel
	}
	// 检查角色是否存在权限关联
	if permHasRel, err := r.RoleRepo.CheckRoleRelPerm(id); err != nil || permHasRel {
		return model.PermissionHasRel
	}

	// 调用数据层删除角色
	return r.RoleRepo.DeleteRole(id)
}
