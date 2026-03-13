package model

import (
	"time"

	"github.com/google/uuid"
)

type SysUser struct {
	UserID         uuid.UUID  `gorm:"column:user_id;primaryKey"`
	Username       string     `gorm:"column:username"`
	PasswordHash   string     `gorm:"column:password_hash"`
	Email          string     `gorm:"column:email"`
	Phone          string     `gorm:"column:phone"`
	Status         string     `gorm:"column:status"`
	LastLoginTime  *time.Time `gorm:"column:last_login_time`
	LastLoginIP    string     `gorm:"column:last_login_ip`
	FailedAttempts int        `gorm:"column:failed_attempts`
	LockUntil      *time.Time `gorm:"column:lock_until`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

type UserProfile struct {
	UserInfoID uuid.UUID  `gorm:"column:user_info_id;primaryKey"`
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

type RegisterRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}
