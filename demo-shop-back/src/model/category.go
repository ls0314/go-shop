package model

import "time"

// SysCategory 类目结构体, 对应数据表sys_category
type SysCategory struct {
	CategoryId    int64          `gorm:"primaryKey;column:category_id" json:"category_id"`
	ParentId      int64          `gorm:"column:parent_id" json:"parent_id"`
	CategoryName  string         `gorm:"column:category_name" json:"category_name"`
	CategoryLevel int64          `gorm:"column:category_level" json:"category_level"`
	SortOrder     int64          `gorm:"column:sort_order" json:"sort_order"`
	IsLeaf        bool           `gorm:"column:is_leaf" json:"is_leaf"`
	IsVisible     bool           `gorm:"column:is_visible" json:"is_visible"`
	Status        string         `gorm:"column:status" json:"status"`
	CategoryPath  string         `gorm:"column:category_path" json:"category_path"`
	IconUrl       string         `gorm:"column:icon_url" json:"icon_url"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	CreateBy      int64          `gorm:"column:create_by" json:"create_by"`
	UpdateBy      int64          `gorm:"column:update_by" json:"update_by"`
	Children      []*SysCategory `gorm:"-" json:"children"`
}

func (SysCategory) TableName() string {
	return "sys_category"
}
