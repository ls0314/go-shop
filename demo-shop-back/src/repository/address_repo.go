package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AddressRepo 用户地址表数据层实例
type AddressRepo struct {
	DB *gorm.DB // 全局数据库
}

// NewAddressRepo 创建地址表数据层实例
// 接收值：使用全局数据库，故无接收值
// 返回值：*AddressRepo - 地址表数据层指针
func NewAddressRepo() *AddressRepo {
	return &AddressRepo{
		DB: db.DB,
	}
}

// WithTx 切换数据库事务实例
// 接收值：tx - 数据库事务实例
// 返回值：*AddressRepo - 绑定事务的地址表数据层指针
func (ar *AddressRepo) WithTx(tx *gorm.DB) *AddressRepo {
	return &AddressRepo{
		DB: tx,
	}
}

// CreateAddress 创建地址记录
// 接收值：address - 地址对象指针
// 返回值：addressId - 新建地址ID, error - 错误信息
func (ar *AddressRepo) CreateAddress(address *model.UserAddress) (addressId int64, err error) {
	err = ar.DB.Create(&address).Error
	addressId = address.AddressId
	return addressId, err
}

// GetAddress 查询地址信息（按地址ID查，排除已删除）
// 接收值：addressId - 地址唯一标识
// 返回值：*model.UserAddress - 地址对象指针, error - 地址不存在返回AddressNotExist
func (ar *AddressRepo) GetAddress(addressId int64) (address *model.UserAddress, err error) {
	err = ar.DB.Where("address_id = ? AND is_deleted = ?", addressId, false).First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressNotExist
	}
	return address, err
}

// GetDefaultAddress 查询用户当前默认地址（排除已删除）
// 接收值：userId - 用户ID
// 返回值：*model.UserAddress - 默认地址对象指针, error - 无默认地址返回AddressNotExist
func (ar *AddressRepo) GetDefaultAddress(userId int64) (address *model.UserAddress, err error) {
	err = ar.DB.Where("user_id = ? AND is_default = ? AND is_deleted = ?", userId, true, false).First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressNotExist
	}
	return address, err
}

// GetAddressList 查询用户所有有效地址列表（排除已删除，默认地址排最前）
// 接收值：userId - 用户ID
// 返回值：[]model.UserAddress - 地址列表, error - 列表为空返回AddressListIsNull
func (ar *AddressRepo) GetAddressList(userId int64) ([]model.UserAddress, error) {
	var addressList []model.UserAddress
	addrDB := ar.DB.Where("user_id = ? AND is_deleted = ?", userId, false).Order("is_default DESC, updated_at DESC")
	err := addrDB.Find(&addressList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressListIsNull
	}
	return addressList, err
}

// UpdateAddress 更新地址信息（全字段Save，须在事务内调用以保证一致性）
// 接收值：updateAddress - 更新后的地址对象指针
// 返回值：error - 错误信息
func (ar *AddressRepo) UpdateAddress(updateAddress *model.UserAddress) error {
	return ar.DB.Save(updateAddress).Error
}

// DeleteAddress 软删除地址（设置 is_deleted=true，不物理删除）
// 接收值：addressId - 待删除地址ID
// 返回值：error - 错误信息
func (ar *AddressRepo) DeleteAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}

// SetDefaultAddress 将指定地址设为默认地址
// 接收值：addressId - 地址ID
// 返回值：error - 错误信息
func (ar *AddressRepo) SetDefaultAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": true,
			"updated_at": time.Now(),
		}).Error
}

// SetCommonAddress 取消指定地址的默认标记（设为非默认）
// 接收值：addressId - 地址ID
// 返回值：error - 错误信息
func (ar *AddressRepo) SetCommonAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": false,
			"updated_at": time.Now(),
		}).Error
}
