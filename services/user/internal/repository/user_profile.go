package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// UserProfileRepo 用户档案表数据层实例
type UserProfileRepo struct {
	DB *gorm.DB
}

// NewUserProfileRepo 创建用户档案表数据层实例
func NewUserProfileRepo(conn *gorm.DB) *UserProfileRepo {
	return &UserProfileRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (ur *UserProfileRepo) WithTx(tx *gorm.DB) *UserProfileRepo {
	return &UserProfileRepo{DB: tx}
}

// CreateProfile 创建用户档案
func (ur *UserProfileRepo) CreateProfile(profile *model.UserProfile) error {
	return ur.DB.Create(profile).Error
}

// GetProfileByUserId 按用户ID查询档案。查不到返回 gorm.ErrRecordNotFound。
func (ur *UserProfileRepo) GetProfileByUserId(userId int64) (*model.UserProfile, error) {
	var profile model.UserProfile
	if err := ur.DB.Where("user_id = ?", userId).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile 更新用户档案
func (ur *UserProfileRepo) UpdateProfile(profile *model.UserProfile) error {
	return ur.DB.Save(profile).Error
}

// DeleteProfile 按用户ID删除档案
func (ur *UserProfileRepo) DeleteProfile(userId int64) error {
	return ur.DB.Where("user_id = ?", userId).Delete(&model.UserProfile{}).Error
}
