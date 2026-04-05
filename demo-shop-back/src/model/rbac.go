package model

import (
	"time"
)

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

type SysMenu struct {
	MenuId        int64                  `gorm:"primaryKey;column:menu_id" json:"menu_id"`
	ParentId      int64                  `gorm:"column:parent_id" json:"parent_id"`
	MenuName      string                 `gorm:"column:menu_name" json:"menu_name"`
	MenuType      string                 `gorm:"column:menu_type" json:"menu_type"`
	Icon          string                 `gorm:"column:icon" json:"icon"`
	RoutePath     string                 `gorm:"column:route_path" json:"route_path"`
	ComponentPath string                 `gorm:"column:component" json:"component"`
	IsVisible     int                    `gorm:"column:is_visible" json:"is_visible"`
	IsCache       int                    `gorm:"column:is_cache" json:"is_cache"`
	SortOrder     int64                  `gorm:"column:sort_order" json:"sort_order"`
	MetaInfo      map[string]interface{} `gorm:"column:meta_info" json:"meta_info"`
	CreatedAt     time.Time              `gorm:"column:created_at" json:"created_at"`
	Children      []SysMenu              `gorm:"-" json:"children" `
}

func (SysMenu) TableName() string {
	return "sys_menu"
}

type SysRole struct {
	RoleId      int64     `gorm:"primaryKey;column:role_id" json:"role_id"`
	RoleName    string    `gorm:"column:role_name" json:"role_name"`
	RoleType    string    `gorm:"column:role_type" json:"role_type"`
	Description string    `gorm:"column:description" json:"description"`
	IsSystem    bool      `gorm:"column:is_system" json:"is_system"`
	IsDefault   bool      `gorm:"column:is_default" json:"is_default"`
	DataScope   string    `gorm:"column:data_scope" json:"data_scope"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy   string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy   string    `gorm:"column:updated_by" json:"updated_by"`
}

func (SysRole) TableName() string {
	return "sys_role"
}
