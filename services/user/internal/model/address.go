package model

import "time"

// UserAddress 收货地址实体,对应数据表 user_address。
//
// 表在 C1 就按归属建到了 user_db(见 migrations/000006),但读写路径一直
// 留在单体 —— 于是两个库各有一张同名表、其中 user_db 那张是空的。
// 本次把路径搬过来(DS-A-25 §4.5.2 第 1 条:address → user-service)。
//
// **软删除**:is_deleted 而非物理删除。为什么保留它:地址会被历史订单
// 的 address_snapshot 引用过(快照是独立存储的,不是外键),物理删除不会
// 破坏订单,但会让"这条地址什么时候消失的"无法追溯。软删的成本只是每个
// 查询都要带 is_deleted = false 这个条件。
type UserAddress struct {
	AddressId     int64     `gorm:"primaryKey;column:address_id" json:"address_id"`
	UserId        int64     `gorm:"column:user_id" json:"user_id"`
	ReceiverName  string    `gorm:"column:receiver_name" json:"receiver_name"`
	ReceiverPhone string    `gorm:"column:receiver_phone" json:"receiver_phone"`
	Province      string    `gorm:"column:province" json:"province"`
	City          string    `gorm:"column:city" json:"city"`
	District      string    `gorm:"column:district" json:"district"`
	DetailAddress string    `gorm:"column:detail_address" json:"detail_address"`
	PostalCode    string    `gorm:"column:postal_code" json:"postal_code"`
	IsDefault     bool      `gorm:"column:is_default" json:"is_default"`
	AddressTag    string    `gorm:"column:address_tag" json:"address_tag"`
	IsDeleted     bool      `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserAddress) TableName() string {
	return "user_address"
}

// AddressMaxCount 单个用户的地址数量上限。
//
// 与单体的 20 条一致。上限的意义不是"数据库撑不住",而是
// **结算页的地址选择器**:超过一屏就没人翻了,而用户多到那个程度
// 说明他该整理地址簿而不是继续加。
const AddressMaxCount = 20
