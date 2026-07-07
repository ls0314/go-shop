package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"errors"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type AddressService struct {
	AddressRepo *repository.AddressRepo
	DB          *gorm.DB
}

func NewAddressService() *AddressService {
	return &AddressService{
		AddressRepo: repository.NewAddressRepo(),
		DB:          db.DB,
	}
}

func (as *AddressService) CreateAddress(address *model.UserAddress) (addressId int64, err error) {
	if !ValidatePhone(address.ReceiverPhone) {
		return 0, model.PhoneMalformed
	}

	err = as.DB.Transaction(func(tx *gorm.DB) error {

		if address.ReceiverName == "" || address.ReceiverPhone == "" {
			return model.ReceiverNotNull
		}

		addressTx := as.AddressRepo.WithTx(tx)

		defaultAddr, err := addressTx.GetDefaultAddress(address.UserId)
		if err != nil && !errors.Is(err, model.AddressNotExist) {
			return err
		}
		if defaultAddr == nil {
			address.IsDefault = true
		} else {
			if address.IsDefault {
				err = addressTx.SetCommonAddress(defaultAddr.AddressId)
				if err != nil {
					return err
				}
			}
		}
		addressList, err := addressTx.GetAddressList(address.UserId)
		if err != nil && !errors.Is(err, model.AddressListIsNull) {
			return err
		}
		if len(addressList) >= 20 {
			return model.AddressNumsIsFull
		}

		addressId, err = addressTx.CreateAddress(address)
		if err != nil {
			return err
		}
		return nil
	})
	return addressId, err
}

func (as *AddressService) GetAddressList(userId int64) ([]*response.AddressResp, error) {
	addrs, err := as.AddressRepo.GetAddressList(userId)
	if err != nil && !errors.Is(err, model.AddressListIsNull) {
		return nil, err
	}

	addrList := make([]*response.AddressResp, 0, len(addrs))
	for _, addr := range addrs {
		addrList = append(addrList, &response.AddressResp{
			AddressId:     addr.AddressId,
			ReceiverName:  addr.ReceiverName,
			ReceiverPhone: addr.ReceiverPhone,
			Province:      addr.Province,
			City:          addr.City,
			District:      addr.District,
			DetailAddress: addr.DetailAddress,
			PostalCode:    addr.PostalCode,
			IsDefault:     addr.IsDefault,
			AddressTag:    addr.AddressTag,
			CreatedAt:     addr.CreatedAt.String(),
		})
	}
	return addrList, nil

}

func (as *AddressService) GetAddress(userId, addressId int64) (*response.AddressResp, error) {
	address, err := as.AddressRepo.GetAddress(addressId)
	if err != nil {
		return nil, err
	}
	if userId != address.UserId {
		return nil, model.UserNotSetAddress
	}
	addressResp := &response.AddressResp{
		AddressId:     address.AddressId,
		ReceiverName:  address.ReceiverName,
		ReceiverPhone: address.ReceiverPhone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		PostalCode:    address.PostalCode,
		IsDefault:     address.IsDefault,
		AddressTag:    address.AddressTag,
		CreatedAt:     address.CreatedAt.String(),
	}
	return addressResp, nil
}

func (as *AddressService) UpdateAddress(userId, addressId int64, updateAddress map[string]interface{}) (*response.AddressResp, error) {

	err := as.DB.Transaction(func(tx *gorm.DB) error {

		addressTx := as.AddressRepo.WithTx(tx)

		oldAddress, err := addressTx.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != oldAddress.UserId {
			return model.UserNotSetAddress
		}

		newAddress := *oldAddress

		config := &mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &newAddress,
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return err
		}
		if err := decoder.Decode(updateAddress); err != nil {
			return err
		}

		if newAddress.IsDefault != oldAddress.IsDefault && newAddress.IsDefault {
			defaultAddr, err := addressTx.GetDefaultAddress(oldAddress.UserId)
			if err != nil && !errors.Is(err, model.AddressNotExist) {
				return err
			}
			if defaultAddr != nil {
				err = addressTx.SetCommonAddress(defaultAddr.AddressId)
				if err != nil {
					return err
				}
			}
		}

		err = addressTx.UpdateAddress(&newAddress)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return as.GetAddress(userId, addressId)
}

func (as *AddressService) DeleteAddress(userId, addressId int64) error {
	return as.DB.Transaction(func(tx *gorm.DB) error {
		addressTx := as.AddressRepo.WithTx(tx)

		address, err := addressTx.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != address.UserId {
			return model.UserNotSetAddress
		}

		err = addressTx.DeleteAddress(addressId)
		if err != nil {
			return err
		}
		// TODO:订单校验

		if address.IsDefault {
			addressList, err := addressTx.GetAddressList(userId)
			if err != nil {
				return err
			}
			if len(addressList) == 0 {
				return model.AddressNotExist
			}
			err = addressTx.SetDefaultAddress(addressList[0].AddressId)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (as *AddressService) SetDefaultAddress(userId, addressId int64) error {
	return as.DB.Transaction(func(tx *gorm.DB) error {

		address, err := as.AddressRepo.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != address.UserId {
			return model.UserNotSetAddress
		}

		addressTx := as.AddressRepo.WithTx(tx)
		defaultAddr, err := addressTx.GetDefaultAddress(userId)
		if err != nil && !errors.Is(err, model.AddressNotExist) {
			return err
		}
		if defaultAddr != nil {
			err = addressTx.SetCommonAddress(defaultAddr.AddressId)
			if err != nil {
				return err
			}
		}
		err = addressTx.SetDefaultAddress(addressId)
		if err != nil {
			return err
		}
		return nil
	})
}
