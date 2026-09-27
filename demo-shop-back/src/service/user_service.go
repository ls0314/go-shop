package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
	"regexp"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ValidatePassword 密码格式校验
// 规则：长度≥8，包含字母、数字、特殊符号
// 接收值：password - 明文密码
// 返回值：bool - 校验结果
func ValidatePassword(password string) bool {
	// 校验密码长度
	if len(password) < 8 {
		return false
	}
	// 校验是否包含字母
	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	// 校验是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 校验是否包含特殊符号
	hasSymbol := regexp.MustCompile(`[^\w\s]`).MatchString(password)

	return hasLetter && hasNumber && hasSymbol
}

// ValidatePhone 手机号格式校验
// 规则：中国大陆11位手机号
// 接收值：phone - 手机号
// 返回值：bool - 校验结果
func ValidatePhone(phone string) bool {
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(phone)
	return reg
}

// recordLoginLog 记录用户登录日志
// 接收值：userID - 用户ID, ip - 登录IP, device - 登录设备, status - 登录状态, reason - 失败原因
func recordLoginLog(userID int64, ip string, device string, status string, reason string) {
	// 执行SQL插入登录日志
	db.DB.Exec(`
	INSERT INTO user_login_log
	(user_id, login_ip, login_device, login_status, failure_reason)
	VALUES (?, ?, ?, ?, ?)
	`, userID, ip, device, status, reason)
}

// UserService 用户服务层实例
type UserService struct {
	UserRepo     *repository.UserRepo        // 用户表数据层实例
	UserInfoRepo *repository.UserProfileRepo // 用户档案表数据层实例
	db           *gorm.DB                    // 全局数据库
}

// NewUserService 创建用户服务层实例
// 接收值：conn - 数据库连接（由调用方注入）
// 返回值：*UserService - 用户服务层指针
func NewUserService() *UserService {
	return &UserService{
		UserRepo:     repository.NewUserRepo(db.DB),
		UserInfoRepo: repository.NewUserProfileRepo(db.DB),
		db:           db.DB,
	}
}

// GetUser 根据用户ID查询用户信息
// 接收值：userId - 用户ID
// 返回值：*model.SysUser - 用户对象，error - 错误信息
func (u *UserService) GetUser(userId int64) (*model.SysUser, error) {
	// 根据用户ID查询用户
	user, err := u.UserRepo.GetUserById(userId)
	if err != nil || user == nil {
		return nil, model.UserNotExist
	}
	return user, nil
}

func (u *UserService) GetUserList(page, pageSize int, status string) ([]model.SysUser, int64, error) {
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
	return u.UserRepo.GetUserList(page, pageSize, status)
}

// UpdateUser 更新用户信息
// 接收值：userId - 用户ID，user - 待更新字段map
// 返回值：error - 错误信息
func (u *UserService) UpdateUser(userId int64, user map[string]interface{}) error {
	// 查询原用户信息，校验是否存在
	oleUser, err := u.UserRepo.GetUserById(userId)
	if err != nil || oleUser == nil {
		return model.UserNotExist
	}

	// 复制原用户数据，用于接收更新字段
	newUser := *oleUser

	// 配置mapstructure解码规则
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newUser,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	// 将更新参数映射到用户对象
	if err := decoder.Decode(user); err != nil {
		return err
	}

	// 校验密码格式
	if !ValidatePassword(newUser.PasswordHash) {
		return model.RegPasswordInvalid
	}
	// 校验手机号格式
	if !ValidatePhone(newUser.Phone) {
		return model.PhoneMalformed
	}

	// 校验用户名是否重复
	if userNameExist, err := u.UserRepo.GetUserByName(newUser.Username); err != nil || userNameExist != nil {
		return model.UsernameExist
	}
	// 校验手机号是否重复
	if userPhoneExist, err := u.UserRepo.GetUserByPhone(newUser.Phone); err != nil || userPhoneExist != nil {
		return model.PhoneExist
	}
	// 校验邮箱是否重复
	if userEmailExist, err := u.UserRepo.GetUserByEmail(newUser.Email); err != nil || userEmailExist != nil {
		return model.EmailExist
	}

	// 密码加密处理
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.PasswordHash = string(passwordHash)

	// 调用数据层更新用户信息
	return u.UserRepo.UpdateUser(&newUser)
}

// DeleteUser 删除用户
// 接收值：userId - 用户ID
// 返回值：error - 错误信息
func (u *UserService) DeleteUser(userId int64) error {
	// 校验用户是否存在
	user, err := u.UserRepo.GetUserById(userId)
	if err != nil || user == nil {
		return model.UserNotExist
	}
	// 检查用户是否存在部门关联
	if deptHasRel, err := u.UserRepo.CheckUserRelDept(userId); err != nil || deptHasRel {
		return model.DeptHasRel
	}
	// 检查用户是否存在角色关联
	if RoleHasRel, err := u.UserRepo.CheckUserRelRole(userId); err != nil || RoleHasRel {
		return model.RoleHasRel
	}

	// 调用数据层删除用户
	return u.UserRepo.DeleteUser(userId)
}

// CreateUser 用户注册
// 事务保证：用户+档案数据一致性
// 接收值：user - 用户注册对象
// 返回值：error - 错误信息
func (u *UserService) CreateUser(user *model.SysUser) error {
	// 校验密码格式
	if !ValidatePassword(user.PasswordHash) {
		return model.RegPasswordInvalid
	}
	// 校验手机号格式
	if !ValidatePhone(user.Phone) {
		return model.PhoneMalformed
	}

	// 开启事务
	tx := u.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 实例化事务绑定的仓库
	userTxRepo := u.UserRepo.WithTx(tx)
	userInfoTxRepo := u.UserInfoRepo.WithTx(tx)

	// 校验用户名是否重复
	if userExisting, err := userTxRepo.GetUserByName(user.Username); err != nil || userExisting != nil {
		tx.Rollback()
		return model.UsernameExist
	}
	// 校验手机号是否重复
	if phoneExisting, err := userTxRepo.GetUserByPhone(user.Phone); err != nil || phoneExisting != nil {
		tx.Rollback()
		return model.PhoneExist
	}
	// 校验邮箱是否重复
	if emailExisting, err := userTxRepo.GetUserByEmail(user.Email); err != nil || emailExisting != nil {
		tx.Rollback()
		return model.EmailExist
	}

	// 密码加密处理
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return err
	}
	user.PasswordHash = string(passwordHash)

	// 创建用户
	if err := userTxRepo.CreateUser(user); err != nil {
		tx.Rollback()
		return err
	}
	// 初始化用户默认档案
	profile := &model.UserProfile{
		UserId:   user.UserID, // 绑定用户ID
		Nickname: "默认用户名",     // 默认初始化
	}

	// 创建用户档案
	if err := userInfoTxRepo.CreateUserProfile(profile); err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error

}

// Login 用户登录
// 支持用户名/手机号登录，生成JWT令牌
// 接收值：user - 登录请求对象, ip - 登录IP, device - 登录设备
// 返回值：*model.LoginResponse - 登录响应，error - 错误信息
func (u *UserService) Login(user *model.LoginRequest, ip string, device string) (*model.LoginResponse, error) {
	var userExist *model.SysUser
	var err error
	// 根据用户名查询用户
	if user.Username != "" {
		userExist, err = u.UserRepo.GetUserByName(user.Username)
		if err != nil || userExist == nil {
			recordLoginLog(0, ip, device, "fail", "user not found")
			return nil, model.UserNotExist
		}
	} else if user.Phone != "" {
		// 根据手机号查询用户
		userExist, err = u.UserRepo.GetUserByPhone(user.Phone)
		if err != nil || userExist == nil {
			recordLoginLog(0, ip, device, "fail", "user not found")
			return nil, model.UserNotExist
		}
	}

	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(userExist.PasswordHash), []byte(user.Password))
	if err != nil {
		return nil, model.LoginPasswordInvalid
	}

	// 记录登录成功日志
	recordLoginLog(userExist.UserID, ip, device, "success", "")

	// 生成JWT令牌
	jwtService := middleware.GetJWTService()
	accessToken, err := jwtService.GenerateAccessToken(userExist.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtService.GenerateRefreshToken(userExist.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	// 返回登录响应
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}
