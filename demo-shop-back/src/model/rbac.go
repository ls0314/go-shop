package model

import (
	"time"

	"gorm.io/datatypes"
)

// SysPermission 权限信息结构体, 对应数据表sys_permission
type SysPermission struct {
	PermissionID   int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	PermissionCode string    `gorm:"column:permission_code;uniqueIndex;not null;size:100" json:"permission_code"`
	PermissionName string    `gorm:"column:permission_name;not null;size:100" json:"permission_name"`
	PermissionType string    `gorm:"column:permission_type;not null;size:20" json:"permission_type"`
	RequestMethod  string    `gorm:"column:request_method;size:10" json:"request_method"`
	ApiPath        string    `gorm:"column:api_path;not null;size:500" json:"api_path"`
	Description    string    `gorm:"column:description;type:text" json:"description"`
	IsSystem       bool      `gorm:"column:is_system;default:false" json:"is_system"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (SysPermission) TableName() string {
	return "sys_permission"
}

// SysMenu 菜单信息结构体, 对应数据表sys_menu
type SysMenu struct {
	MenuId        int64             `gorm:"primaryKey;column:menu_id" json:"menu_id"`
	ParentId      int64             `gorm:"column:parent_id" json:"parent_id"`
	MenuName      string            `gorm:"column:menu_name" json:"menu_name"`
	MenuType      string            `gorm:"column:menu_type" json:"menu_type"`
	Icon          string            `gorm:"column:icon" json:"icon"`
	RoutePath     string            `gorm:"column:route_path" json:"route_path"`
	ComponentPath string            `gorm:"column:component" json:"component"`
	IsVisible     bool              `gorm:"column:is_visible" json:"is_visible"`
	IsCache       bool              `gorm:"column:is_cache" json:"is_cache"`
	SortOrder     int64             `gorm:"column:sort_order" json:"sort_order"`
	MetaInfo      datatypes.JSONMap `gorm:"column:meta_info;type:jsonb" json:"meta_info"`
	CreatedAt     time.Time         `gorm:"column:created_at" json:"created_at"`
	Children      []SysMenu         `gorm:"-" json:"children" `
}

func (SysMenu) TableName() string {
	return "sys_menu"
}

// SysRole 角色信息结构体, 对应数据表sys_role
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
	CreatedBy   int64     `gorm:"column:create_by" json:"create_by"`
	UpdatedBy   int64     `gorm:"column:update_by" json:"update_by"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type SysDept struct {
	DeptId    int64     `gorm:"primaryKey;column:dept_id" json:"dept_id"`
	DeptName  string    `gorm:"column:dept_name" json:"dept_name"`
	ParentId  int64     `gorm:"column:parent_id" json:"parent_id"`
	DeptType  string    `gorm:"column:dept_type" json:"dept_type"`
	LeaderId  int64     `gorm:"column:leader_id" json:"leader_id"`
	SortOrder int64     `gorm:"column:sort_order" json:"sort_order"`
	Status    string    `gorm:"column:status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SysDept) TableName() string {
	return "sys_department"
}

type SysScope struct {
	ScopeId        int64     `gorm:"primaryKey;column:scope_id" json:"scope_id"`
	RoleId         int64     `gorm:"column:role_id" json:"role_id"`
	ResourceType   string    `gorm:"column:resource_type" json:"resource_type"`
	FieldName      string    `gorm:"column:field_name" json:"field_name"`
	ConditionType  string    `gorm:"column:condition_type" json:"condition_type"`
	ConditionValue string    `gorm:"column:condition_value" json:"condition_value"`
	Description    string    `gorm:"column:description" json:"description"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SysScope) TableName() string { return "sys_data_scope" }
