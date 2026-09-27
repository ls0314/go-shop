package model

import (
	"time"
)

// SysUser 用户注册结构体, 对应数据表sys_user
type SysUser struct {
	UserID         int64      `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username       string     `gorm:"column:username" json:"username"`
	PasswordHash   string     `gorm:"column:password_hash" json:"password"`
	Email          string     `gorm:"column:email" json:"email"`
	Phone          string     `gorm:"column:phone" json:"phone"`
	Status         string     `gorm:"column:status" json:"status"`
	FailedAttempts int        `gorm:"column:failed_attempts" json:"failed_attempts"`
	LockUntil      *time.Time `gorm:"column:lock_until" json:"lock_until"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

// UserProfile 用户信息结构体, 对应数据表user_profile
type UserProfile struct {
	UserInfoID int64      `gorm:"column:user_info_id;primaryKey" json:"user_info_id"`
	UserId     int64      `gorm:"column:user_id" json:"user_id"`
	Nickname   string     `gorm:"column:nickname" json:"nickname"`
	RealName   string     `gorm:"column:real_name" json:"real_name"`
	Gender     string     `gorm:"column:gender" json:"gender"`
	AvatarURL  string     `gorm:"column:avatar_url" json:"avatar_url"`
	Birthdate  *time.Time `gorm:"column:birthdate" json:"birthdate"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (UserProfile) TableName() string {
	return "user_profile"
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Phone    string `json:"phone"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 签发Token结构
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserLoginInfo struct {
	UserID       int64
	Username     string
	PasswordHash string
}

// UserLoginLog 登录日志(B0 去全局化时从 service 里的裸 SQL 收敛为模型)。
//
// 只声明写入路径需要的字段:login_time / login_type / location / created_at
// 均由 DB 默认值填充(见 000001 迁移),无需在 Go 侧赋值 —— 这样也避免了
// "Go 侧零值覆盖 DB 默认值"的问题。
type UserLoginLog struct {
	Id            int64  `gorm:"column:id;primary_key" json:"id"`
	UserId        int64  `gorm:"column:user_id" json:"user_id"`
	LoginIp       string `gorm:"column:login_ip" json:"login_ip"`
	LoginDevice   string `gorm:"column:login_device" json:"login_device"`
	LoginStatus   string `gorm:"column:login_status" json:"login_status"`
	FailureReason string `gorm:"column:failure_reason" json:"failure_reason"`
}

func (UserLoginLog) TableName() string { return "user_login_log" }
