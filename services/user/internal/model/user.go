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
