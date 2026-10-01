package model

import "time"

// SysUser 用户结构体, 对应数据表 sys_user。
// 当前只服务绑定域的主键校验,用户域迁入时再补 UserProfile / UserLoginLog。
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

// UserProfile 用户档案,对应数据表 user_profile
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

func (UserProfile) TableName() string { return "user_profile" }

// UserLoginLog 登录日志,对应数据表 user_login_log
type UserLoginLog struct {
	Id            int64  `gorm:"column:id;primary_key" json:"id"`
	UserId        int64  `gorm:"column:user_id" json:"user_id"`
	LoginIp       string `gorm:"column:login_ip" json:"login_ip"`
	LoginDevice   string `gorm:"column:login_device" json:"login_device"`
	LoginStatus   string `gorm:"column:login_status" json:"login_status"`
	FailureReason string `gorm:"column:failure_reason" json:"failure_reason"`
}

func (UserLoginLog) TableName() string { return "user_login_log" }
