package repository

import (
	"time"

	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// AddressRepo 收货地址数据层。
//
// 与单体 address_repo.go 的差异只有一处:查不到时**透传
// gorm.ErrRecordNotFound,不在这里翻译成业务错误** —— 与 user / product /
// marketing 三个服务的既有口径一致(仓储不判业务语义,由 logic 判)。
// 单体那份在仓储里就返回 AddressNotExist,会让"表里没有"与"业务上不可用"
// 这两件事在分层上混在一起。
type AddressRepo struct {
	DB *gorm.DB
}

func NewAddressRepo(conn *gorm.DB) *AddressRepo {
	return &AddressRepo{DB: conn}
}

func (a *AddressRepo) WithTx(tx *gorm.DB) *AddressRepo {
	return &AddressRepo{DB: tx}
}

// CreateAddress 新建地址,回填自增主键
func (a *AddressRepo) CreateAddress(addr *model.UserAddress) error {
	return a.DB.Create(addr).Error
}

// GetAddressById 按主键查(排除已软删)
func (a *AddressRepo) GetAddressById(addressId int64) (*model.UserAddress, error) {
	var addr model.UserAddress
	err := a.DB.Where("address_id = ? AND is_deleted = ?", addressId, false).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// GetDefaultAddress 查用户的默认地址。
//
// 返回 gorm.ErrRecordNotFound 表示"没有默认地址"—— 那是**正常状态**
// (比如用户删掉了最后一条地址),不是错误。调用方据此决定"要不要把
// 新地址设成默认",故不能用业务错误把它盖掉。
func (a *AddressRepo) GetDefaultAddress(userId int64) (*model.UserAddress, error) {
	var addr model.UserAddress
	err := a.DB.Where("user_id = ? AND is_default = ? AND is_deleted = ?", userId, true, false).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// GetAddressList 查用户全部有效地址。
//
// 排序 `is_default DESC, updated_at DESC`:默认地址排最前,其余按最近更新。
// 这个顺序**同时被用在两个地方** —— 地址列表页的展示顺序,以及
// "删除默认地址后顶上来的那条"(取列表第一条)。所以它不是纯展示偏好,
// 改它会影响业务行为。
func (a *AddressRepo) GetAddressList(userId int64) ([]*model.UserAddress, error) {
	var list []*model.UserAddress
	err := a.DB.Where("user_id = ? AND is_deleted = ?", userId, false).
		Order("is_default DESC, updated_at DESC").
		Find(&list).Error
	return list, err
}

// CountAddress 统计用户有效地址条数(上限校验用)。
//
// 单独一个 COUNT 而不是用 GetAddressList 的 len:上限校验只需要一个数,
// 为此把整张地址簿拉进内存没有必要 —— 而地址条数是有上限的(20),
// 所以这条不是性能问题,是"别让一个校验动作依赖一个查询的顺序与过滤条件"。
func (a *AddressRepo) CountAddress(userId int64) (int64, error) {
	var n int64
	err := a.DB.Model(&model.UserAddress{}).
		Where("user_id = ? AND is_deleted = ?", userId, false).
		Count(&n).Error
	return n, err
}

// UpdateAddress 全字段保存(须在事务内调用)
func (a *AddressRepo) UpdateAddress(addr *model.UserAddress) error {
	return a.DB.Save(addr).Error
}

// DeleteAddress 软删除
func (a *AddressRepo) DeleteAddress(addressId int64) error {
	return a.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}

// SetDefault 设为默认
func (a *AddressRepo) SetDefault(addressId int64) error {
	return a.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": true,
			"updated_at": time.Now(),
		}).Error
}

// UnsetDefault 取消默认标记
func (a *AddressRepo) UnsetDefault(addressId int64) error {
	return a.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": false,
			"updated_at": time.Now(),
		}).Error
}
