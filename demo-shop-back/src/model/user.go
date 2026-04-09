package model

import (
	"time"
)

// SysUser 用户注册结构体, 对应数据表sys_user
type SysUser struct {
	UserID         int64      `gorm:"column:user_id;primaryKey"`
	Username       string     `gorm:"column:username"`
	PasswordHash   string     `gorm:"column:password_hash"`
	Email          string     `gorm:"column:email"`
	Phone          string     `gorm:"column:phone"`
	Status         string     `gorm:"column:status"`
	FailedAttempts int        `gorm:"column:failed_attempts"`
	LockUntil      *time.Time `gorm:"column:lock_until"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

// UserProfile 用户信息结构体, 对应数据表user_profile
type UserProfile struct {
	UserInfoID int64      `gorm:"column:user_info_id;primaryKey"`
	Nickname   string     `gorm:"column:nickname"`
	RealName   string     `gorm:"column:real_name"`
	Gender     string     `gorm:"column:gender"`
	AvatarURL  string     `gorm:"column:avatar_url"`
	Birthdate  *time.Time `gorm:"column:birthdate"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
