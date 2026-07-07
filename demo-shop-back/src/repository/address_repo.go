package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AddressRepo struct {
	DB *gorm.DB
}

func NewAddressRepo() *AddressRepo {
	return &AddressRepo{
		DB: db.DB,
	}
}

func (ar *AddressRepo) WithTx(tx *gorm.DB) *AddressRepo {
	return &AddressRepo{
		DB: tx,
	}
}

func (ar *AddressRepo) CreateAddress(address *model.UserAddress) (addressId int64, err error) {
	err = ar.DB.Create(&address).Error
	addressId = address.AddressId
	return addressId, err
}

func (ar *AddressRepo) GetAddress(addressId int64) (address *model.UserAddress, err error) {
	err = ar.DB.Where("address_id = ? AND is_deleted = ?", addressId, false).First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressNotExist
	}
	return address, err
}

func (ar *AddressRepo) GetDefaultAddress(userId int64) (address *model.UserAddress, err error) {
	err = ar.DB.Where("user_id = ? AND is_default = ? AND is_deleted = ?", userId, true, false).First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressNotExist
	}
	return address, err
}

func (ar *AddressRepo) GetAddressList(userId int64) ([]model.UserAddress, error) {
	var addressList []model.UserAddress
	addrDB := ar.DB.Where("user_id = ? AND is_deleted = ?", userId, false).Order("is_default DESC, updated_at DESC")
	err := addrDB.Find(&addressList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.AddressListIsNull
	}

	return addressList, err
}

func (ar *AddressRepo) UpdateAddress(updateAddress *model.UserAddress) error {
	return ar.DB.Save(updateAddress).Error
}

func (ar *AddressRepo) DeleteAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_at": time.Now(),
		}).Error
}

func (ar *AddressRepo) SetDefaultAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": true,
			"updated_at": time.Now(),
		}).Error
}

func (ar *AddressRepo) SetCommonAddress(addressId int64) error {
	return ar.DB.Model(&model.UserAddress{}).
		Where("address_id = ?", addressId).
		Updates(map[string]interface{}{
			"is_default": false,
			"updated_at": time.Now(),
		}).Error
}
