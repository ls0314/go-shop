package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// RoleService 角色对象服务层实例
type RoleService struct {
	RoleRepo *repository.RoleRepo // 数据层角色对象指针
	db       *gorm.DB
}

// NewRoleService 新建服务层角色对象实例
// 接收值：roleRepo - 数据层角色对象指针
// 返回值：*RoleService - 服务层角色对象指针
func NewRoleService() *RoleService {
	return &RoleService{
		RoleRepo: repository.NewRoleRepo(db.DB),
		db:       db.DB,
	}
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
	// 开启事务
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	roleTxRepo := r.RoleRepo.WithTx(tx)
	// 调用数据层删除角色
	if err := roleTxRepo.CreateRole(role); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
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
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
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

	// 开启事务
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	roleTxRepo := r.RoleRepo.WithTx(tx)
	// 调用数据层更新角色
	if err := roleTxRepo.UpdateRole(&newRole); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error
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
	// 开启事务
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	roleTxRepo := r.RoleRepo.WithTx(tx)
	// 调用数据层删除角色
	if err := roleTxRepo.DeleteRole(id); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	return tx.Commit().Error

}
