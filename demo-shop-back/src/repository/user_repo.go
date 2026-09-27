package repository

import (
	"demo-shop-back/src/model"
	"errors"

	"gorm.io/gorm"
)

// UserRepo 用户表数据层实例
type UserRepo struct {
	DB *gorm.DB
}

// NewUserRepo 创建用户表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserRepo - 用户表数据层指针
func NewUserRepo(conn *gorm.DB) *UserRepo {
	return &UserRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*UserRepo - 绑定事务的用户表数据层指针
func (ur *UserRepo) WithTx(tx *gorm.DB) *UserRepo {
	return &UserRepo{DB: tx}
}

// CreateUser 创建用户
// 接收值：user - 用户对象指针
// 返回值：error - 错误信息
func (ur *UserRepo) CreateUser(user *model.SysUser) error {
	return ur.DB.Create(user).Error
}

// GetUserById 根据ID查询用户信息
// 接收值：id - 用户唯一标识
// 返回值：*model.SysUser - 用户对象指针，error - 错误信息
func (ur *UserRepo) GetUserById(id int64) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByName 根据用户名查询用户信息
// 接收值：name - 用户名
// 返回值：*model.SysUser - 用户对象指针，error - 错误信息
func (ur *UserRepo) GetUserByName(name string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("username = ?", name).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号查询用户信息
// 接收值：phone - 手机号
// 返回值：*model.SysUser - 用户对象指针，error - 错误信息
func (ur *UserRepo) GetUserByPhone(phone string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱查询用户信息
// 接收值：email - 邮箱
// 返回值：*model.SysUser - 用户对象指针，error - 错误信息
func (ur *UserRepo) GetUserByEmail(email string) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserList 分页查询用户列表
// 接收值：page - 页数，pageSize - 每页条数
// 返回值：[]model.SysUser - 用户列表，int64 - 总条数，error - 错误信息
func (ur *UserRepo) GetUserList(page, pageSize int, status string) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	userDb := ur.DB.Model(&model.SysUser{})

	if status != "" {
		userDb = userDb.Where("status = ?", status)
	}

	if err := userDb.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	if err := userDb.Order("user_id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
// 接收值：user - 用户对象指针
// 返回值：error - 错误信息
func (ur *UserRepo) UpdateUser(user *model.SysUser) error {
	return ur.DB.Save(user).Error
}

// DeleteUser 删除用户
// 接收值：id - 待删除用户唯一标识
// 返回值：error - 错误信息
func (ur *UserRepo) DeleteUser(id int64) error {
	return ur.DB.Delete(&model.SysUser{}, id).Error
}

// InsertLoginLog 写入登录日志
// 接收值：log - 登录日志(login_time/created_at 等由 DB 默认值填充)
// 返回值：error - 错误信息
//
// 为什么放在 repo 层:B0 去全局化之前,这段 INSERT 是 service 里直接 db.DB.Exec 的裸 SQL,
// 既绕过了数据层、又把全局连接耦合进了业务代码。收敛到这里后 service 只依赖 UserRepo。
func (ur *UserRepo) InsertLoginLog(log *model.UserLoginLog) error {
	return ur.DB.Select(
		"user_id", "login_ip", "login_device", "login_status", "failure_reason",
	).Create(log).Error
}

// CheckUserRelDept 检查用户是否关联部门
// 接收值：userId - 用户唯一标识
// 返回值：bool - 是否关联部门，error - 错误信息
func (ur *UserRepo) CheckUserRelDept(userId int64) (bool, error) {
	var count int64
	err := ur.DB.Table("sys_dept_user").Where("user_id=?", userId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckUserRelRole 检查用户是否关联角色
// 接收值：userId - 用户唯一标识
// 返回值：bool - 是否关联角色，error - 错误信息
func (ur *UserRepo) CheckUserRelRole(userId int64) (bool, error) {
	var count int64
	err := ur.DB.Table("sys_role_user").Where("user_id=?", userId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
