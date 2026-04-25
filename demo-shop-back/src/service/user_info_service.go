package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
)

// UserInfoService 用户信息服务层实例
type UserInfoService struct {
	UserInfoRepo *repository.UserProfileRepo // 用户信息表数据层实例
}

// NewUserInfoService 创建用户信息服务层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*UserInfoService - 用户信息服务层指针
func NewUserInfoService() *UserInfoService {
	return &UserInfoService{
		UserInfoRepo: repository.NewUserProfileRepo(),
	}
}

// CreateUserInfo 创建用户信息
// 接收值：userInfo - 用户信息对象
// 返回值：error - 错误信息
func (uif *UserInfoService) CreateUserInfo(userInfo *model.UserProfile) error {
	// 判断用户信息是否存在
	if userInfoExist, err := uif.UserInfoRepo.GetUserProfileByUserId(userInfo.UserId); err != nil || userInfoExist != nil {
		return model.UsernameExist
	}
	// 调用数据层创建用户信息
	return uif.UserInfoRepo.CreateUserProfile(userInfo)
}

// GetUserInfoByUserId 根据用户ID查询用户信息
// 接收值：userId - 用户ID
// 返回值：*model.UserProfile - 用户信息对象，error - 错误信息
func (uif *UserInfoService) GetUserInfoByUserId(userId int64) (*model.UserProfile, error) {
	var userInfo *model.UserProfile
	// 根据用户Id获取用户信息信息
	userInfo, err := uif.UserInfoRepo.GetUserProfileByUserId(userId)
	// 校验查询结果，不存在则返回用户不存在错误
	if err != nil || userInfo == nil {
		return nil, model.UserNotExist
	}
	return userInfo, nil
}

// UpdateUserProfile 更新用户信息信息
// 接收值：userInfoId - 用户ID，updateUserInfo - 待更新字段map
// 返回值：error - 错误信息
func (uif *UserInfoService) UpdateUserProfile(userInfoId int64, updateUserInfo map[string]interface{}) error {
	// 查询原用户信息信息，校验是否存在
	oldUserInfo, err := uif.UserInfoRepo.GetUserProfileByUserId(userInfoId)
	if err != nil || oldUserInfo == nil {
		return model.UserNotExist
	}

	// 复制原信息数据，用于接收更新字段
	newUserInfo := *oldUserInfo
	// 配置mapstructure解码规则
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newUserInfo,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	// 将更新参数映射到用户信息对象
	if err := decoder.Decode(updateUserInfo); err != nil {
		return err
	}

	// 校验禁止修改用户ID
	if oldUserInfo.UserId != newUserInfo.UserId {
		return model.UserIdIsSystem
	}
	// 调用数据层更新用户信息
	return uif.UserInfoRepo.UpdateUserProfile(&newUserInfo)
}

// DeleteUserInfo 根据用户ID删除用户信息
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (uif *UserInfoService) DeleteUserInfo(userId int64) error {
	// 校验用户信息是否存在，不存在则返回错误
	exist, err := uif.UserInfoRepo.GetUserProfileByUserId(userId)
	if err != nil || exist == nil {
		return model.UserNotExist
	}
	// 调用数据层删除用户信息
	return uif.UserInfoRepo.DeleteUserProfile(userId)
}
