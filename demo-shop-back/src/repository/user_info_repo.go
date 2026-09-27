package repository

import (
	"demo-shop-back/src/model"

	"gorm.io/gorm"
)

// UserProfileRepo 用户信息表数据层实例
type UserProfileRepo struct {
	DB *gorm.DB
}

// NewUserProfileRepo 创建用户信息表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserProfileRepo - 用户信息表数据层指针
func NewUserProfileRepo(conn *gorm.DB) *UserProfileRepo {
	return &UserProfileRepo{DB: conn}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*UserProfileRepo - 绑定事务的用户信息表数据层指针
func (ur *UserProfileRepo) WithTx(tx *gorm.DB) *UserProfileRepo {
	return &UserProfileRepo{DB: tx}
}

// CreateUserProfile 创建用户信息
// 接收值：user - 用户信息对象指针
// 返回值：error - 错误信息
func (ur *UserProfileRepo) CreateUserProfile(user *model.UserProfile) error {
	return ur.DB.Create(user).Error
}

// GetUserProfileByUserId 根据用户ID查询用户信息
// 接收值：userId - 用户ID
// 返回值：*model.UserProfile - 用户信息对象指针，error - 错误信息
func (ur *UserProfileRepo) GetUserProfileByUserId(userId int64) (*model.UserProfile, error) {
	var userProfile model.UserProfile
	err := ur.DB.Where("user_id=?", userId).First(&userProfile).Error
	return &userProfile, err
}

// UpdateUserProfile 更新用户信息
// 接收值：user - 用户信息对象指针
// 返回值：error - 错误信息
func (ur *UserProfileRepo) UpdateUserProfile(user *model.UserProfile) error {
	return ur.DB.Save(user).Error
}

// DeleteUserProfile 根据用户ID删除用户信息
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (ur *UserProfileRepo) DeleteUserProfile(userId int64) error {
	return ur.DB.Where("user_id=?", userId).Delete(&model.UserProfile{}).Error
}

// recordLoginLog 记录登录日志
// 接收值：userInfo - 用户信息对象指针
// 返回值：error - 错误信息
func (ur *UserProfileRepo) recordLoginLog(userInfo *model.UserProfile) error {
	return ur.DB.Create(userInfo).Error
}
