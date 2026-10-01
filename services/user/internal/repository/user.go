package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户表数据层实例。
type UserRepo struct {
	DB *gorm.DB
}

// NewUserRepo 创建用户表数据层实例
func NewUserRepo(conn *gorm.DB) *UserRepo {
	return &UserRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (ur *UserRepo) WithTx(tx *gorm.DB) *UserRepo {
	return &UserRepo{DB: tx}
}

// CreateUser 创建用户
func (ur *UserRepo) CreateUser(user *model.SysUser) error {
	return ur.DB.Create(user).Error
}

// GetUserById 根据ID查询用户
func (ur *UserRepo) GetUserById(id int64) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByName 根据用户名查询用户
func (ur *UserRepo) GetUserByName(name string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("username = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号查询用户
func (ur *UserRepo) GetUserByPhone(phone string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱查询用户
func (ur *UserRepo) GetUserByEmail(email string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserList 分页查询用户,按 user_id 倒序
func (ur *UserRepo) GetUserList(page, pageSize int, status string) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	query := ur.DB.Model(&model.SysUser{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("user_id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// UpdateUser 更新用户
func (ur *UserRepo) UpdateUser(user *model.SysUser) error {
	return ur.DB.Save(user).Error
}

// DeleteUser 删除用户
func (ur *UserRepo) DeleteUser(id int64) error {
	return ur.DB.Delete(&model.SysUser{}, id).Error
}

// InsertLoginLog 写入登录日志。
// 用 Select 限定列:login_time / login_type / location / created_at 由 DB 默认值填充,
// 不限定的话 GORM 会把 Go 侧零值写进去覆盖默认值。
func (ur *UserRepo) InsertLoginLog(log *model.UserLoginLog) error {
	return ur.DB.Select(
		"user_id", "login_ip", "login_device", "login_status", "failure_reason",
	).Create(log).Error
}

// HasDeptRel 检查用户是否关联部门
func (ur *UserRepo) HasDeptRel(userId int64) (bool, error) {
	var count int64
	err := ur.DB.Table("sys_user_dept").Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasRoleRel 检查用户是否关联角色
func (ur *UserRepo) HasRoleRel(userId int64) (bool, error) {
	var count int64
	err := ur.DB.Table("sys_user_role").Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
