package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// RoleRepo 数据层角色对象
type RoleRepo struct {
	DB *gorm.DB
}

// NewRoleRepo 新建数据层角色对象
// 接收值：使用全局数据库, 无参数
// 返回值：*RoleRepo - 数据层角色对象指针
func NewRoleRepo(conn *gorm.DB) *RoleRepo {
	return &RoleRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*RoleRepo - 绑定事务的角色表数据层指针
func (r *RoleRepo) WithTx(tx *gorm.DB) *RoleRepo {
	return &RoleRepo{DB: tx}
}

// CreateRole 创建角色数据层对象实例
// 接收值：role - 角色对象指针
// 返回值：error - 错误信息
func (r *RoleRepo) CreateRole(role *model.SysRole) error {
	return r.DB.Create(role).Error
}

// GetRoleById 查询角色信息（按角色ID查）
// 接收值：
//
//	id - 角色唯一标识
//
// 返回值：
//
//	*model.SysRole - 角色对象指针
//	error - 错误信息
func (r *RoleRepo) GetRoleById(id int64) (*model.SysRole, error) {
	var role model.SysRole
	err := r.DB.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName 查询角色信息（按角色名查）
// 接收值：
//
//	name - 待查角色名
//
// 返回值：
//
//	*model.SysRole - 角色对象指针
//	error - 错误信息
func (r *RoleRepo) GetRoleByName(name string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.DB.Where("role_name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleList 分页查询角色信息
// 接收值：
//
//	page - 页码
//	pageSize - 页面大小
//	roleType - 角色类型
//
// 返回值：
//
//	[]model.SysRole - 角色信息列表
//	int64 - 角色总数
//	error - 错误信息
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

// ListRolesByIds 按角色ID列表批量查询,按 role_id 升序
func (r *RoleRepo) ListRolesByIds(roleIds []int64) ([]*model.SysRole, error) {
	var roles []*model.SysRole
	err := r.DB.Where("role_id IN ?", roleIds).Order("role_id asc").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateRole 更新角色信息
// 接收值：role - 角色对象指针
// 返回值：error - 错误信息
func (r *RoleRepo) UpdateRole(role *model.SysRole) error {
	return r.DB.Save(role).Error
}

// DeleteRole 删除角色信息
// 接收值：id - 待删除角色唯一标识
// 返回值：error - 错误信息
func (r *RoleRepo) DeleteRole(id int64) error {
	return r.DB.Delete(&model.SysRole{}, id).Error
}

// CheckRoleRelUser 检查是否存在用户与此角色关联
// 接收值：
//
//	roleID - 待检查角色唯一标识
//
// 返回值：
//
//	bool - 被检查角色是否存在关联用户
//	error - 错误信息
func (r *RoleRepo) CheckRoleRelUser(roleID int64) (bool, error) {
	var count int64
	err := r.DB.Table("sys_user_role").Where("role_id=?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckRoleRelMenu 检查是否存在菜单与此角色关联
// 接收值：
//
//	roleID - 待检查角色唯一标识
//
// 返回值：
//
//	bool - 被检查角色是否存在关联菜单
//	error - 错误信息
func (r *RoleRepo) CheckRoleRelMenu(roleID int64) (bool, error) {
	var count int64
	err := r.DB.Table("sys_role_menu").Where("role_id=?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckRoleRelPerm 检查是否存在权限与此角色关联
// 接收值：
//
//	roleID - 待检查角色唯一标识
//
// 返回值：
//
//	bool - 被检查角色是否存在关联权限
//	error - 错误信息
func (r *RoleRepo) CheckRoleRelPerm(roleID int64) (bool, error) {
	var count int64
	err := r.DB.Table("sys_role_permission").Where("role_id=?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
