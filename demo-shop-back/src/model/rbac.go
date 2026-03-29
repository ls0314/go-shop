package model

import "time"

// SysPermission 对应数据库表 sys_permission
type SysPermission struct {
	PermissionID   int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	PermissionCode string    `gorm:"column:permission_code;uniqueIndex;not null;size:100" json:"permission_code"` // 唯一权限标识
	PermissionName string    `gorm:"column:permission_name;not null;size:100" json:"permission_name"`
	PermissionType string    `gorm:"column:permission_type;not null;size:20" json:"permission_type"` // button/menu/api
	RequestMethod  string    `gorm:"column:request_method;size:10" json:"request_method"`            // GET/POST/PUT/DELETE
	ApiPath        string    `gorm:"column:api_path;not null;size:500" json:"api_path"`
	Description    string    `gorm:"column:description;type:text" json:"description"`
	IsSystem       bool      `gorm:"column:is_system;default:false" json:"is_system"` // 是否系统内置
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (SysPermission) TableName() string {
	return "sys_permission"
}
