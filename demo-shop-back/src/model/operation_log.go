package model

import "time"

type OperationLog struct {
	LogId         int64     `gorm:"columns:log_id;primary_key;AUTO_INCREMENT" json:"log_id"`
	UserId        int64     `gorm:"column:user_id;" json:"user_id"`
	Username      string    `gorm:"column:username" json:"username"`
	Module        string    `gorm:"column:module" json:"module"`
	Operation     string    `gorm:"column:operation" json:"operation"`
	RequestMethod string    `gorm:"column:request_method" json:"request_method"`
	RequestUrl    string    `gorm:"column:request_url" json:"request_url"`
	RequestParams string    `gorm:"column:request_params" json:"request_params"`
	IpAddress     string    `gorm:"column:ip_address" json:"ip_address"`
	UserAgent     string    `gorm:"column:user_agent" json:"user_agent"`
	ExecuteTime   int64     `gorm:"column:execute_time" json:"execute_time"`
	Status        bool      `gorm:"column:status" json:"status"`
	ErrorMessage  string    `gorm:"column:error_message" json:"error_message"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OperationLog) TableName() string { return "sys_operation_log" }
