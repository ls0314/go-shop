package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// PermissionRepo 权限表数据层实例
type PermissionRepo struct {
	DB *gorm.DB // 全局数据库
}

// NewPermissionRepo 创建权限表数据层实例
// 接收值：全局数据库操作 无接收值
// 返回值：*PermissionRepo - 权限表数据层实例指针
func NewPermissionRepo() *PermissionRepo {
	return &PermissionRepo{DB: db.DB}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*MenuRepo - 绑定事务的权限表数据层指针
func (r *PermissionRepo) WithTx(tx *gorm.DB) *PermissionRepo {
	return &PermissionRepo{DB: tx}
}

// CreatePerm 创建权限
// 接收值：perm - 权限对象指针
// 返回值：error - 错误信息
func (r *PermissionRepo) CreatePerm(perm *model.SysPermission) error {
	return r.DB.Create(perm).Error
}

// GetPermByID 查询权限信息(按权限ID查)
// 接收值：id - 所查询权限唯一标识
// 返回值：
//
//	*model.SysPermission - 所查询权限对象指针
//	error - 错误信息
func (r *PermissionRepo) GetPermByID(id int64) (*model.SysPermission, error) {
	var perm model.SysPermission
	err := r.DB.First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GetPermByCodeUk 查询联合权限信息(按权限代码查)（权限代码，api路径，请求方法）
// 接收值：code - 所查询权限代码
// 返回值：
//
//	*model.SysPermission - 所查询权限信息
//	error - 错误信息
func (r *PermissionRepo) GetPermByCodeUk(apiPath, requestMethod, code string) (*model.SysPermission, error) {
	var perm model.SysPermission
	err := r.DB.Where("permission_code = ? AND request_method = ? AND api_path = ?", code, requestMethod, apiPath).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GetPermCodesByApi 查询联合权限对应代码（api路径，请求方法）
// 接收值：
//
//	apiPath - api路径
//	requestMethod - 请求方法
//
// 返回值：
//
//	*model.SysPermission - 所查询权限信息
//	error - 错误信息
func (r *PermissionRepo) GetPermCodesByApi(apiPath, requestMethod string) ([]string, error) {
	var codes []string
	err := r.DB.Model(&model.SysPermission{}).
		Where("api_path = ? AND request_method = ?", apiPath, requestMethod).
		Pluck("permission_code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// GetPermList 分页查询权限信息（可根据类型查询）
// 接收值：
//
//	page - 页数
//	pageSize - 页大小
//	permType - 权限类型
//
// 返回值:
//
//	[]model.SysPermission - 分页权限信息
//	int64 - 权限总数
//	error - 错误信息
func (r *PermissionRepo) GetPermList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	var perms []model.SysPermission
	var total int64
	// 按类型查询权限信息
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
	// 返回分页权限信息, 权限总数, 错误信息
	return perms, total, nil
}

// UpdatePerm 更新权限信息
// 接收值：perm - 权限对象指针
// 返回值：error - 错误信息
func (r *PermissionRepo) UpdatePerm(perm *model.SysPermission) error {
	return r.DB.Save(perm).Error
}

// DeletePerm 删除权限
// 接收值：id - 待删除权限唯一标识
// 返回值：error - 错误信息
func (r *PermissionRepo) DeletePerm(id int64) error {
	return r.DB.Delete(&model.SysPermission{}, id).Error
}

// CheckRoleRelPerm 检查是否有角色关联该权限
// 接收值：
//
//	permID - 所查询权限唯一标识
//
// 返回值：
//
//	bool - 该权限是否有关联角色
//	error - 错误信息
func (r *PermissionRepo) CheckRoleRelPerm(permID int64) (bool, error) {
	var count int64
	err := r.DB.Table("sys_role_permission").Where("permission_id = ?", permID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PermissionRepo) HasPermission(userId int64, permission string) (bool, error) {
	var count int64

	err := r.DB.Table("sys_user_role AS ur").
		Joins("JOIN sys_role_permission AS rp ON ur.role_id = rp.role_id").
		Joins("JOIN sys_permission AS p ON rp.permission_id = p.permission_id").
		Where("ur.user_id = ? AND p.permission_code = ?", userId, permission).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
