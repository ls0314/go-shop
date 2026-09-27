package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
	"errors"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// AddressService 用户地址管理服务层实例
type AddressService struct {
	AddressRepo *repository.AddressRepo // 地址表数据层实例
	DB          *gorm.DB                // 全局数据库实例（用于开启事务）
}

// NewAddressService 创建地址管理服务层实例
// 接收值：使用全局数据库和repository初始化，故无接收值
// 返回值：*AddressService - 地址服务层指针
func NewAddressService(deps ServiceDeps) *AddressService {
	return &AddressService{
		AddressRepo: repository.NewAddressRepo(deps.DB),
		DB:          deps.DB,
	}
}

// CreateAddress 新增收货地址
// 业务规则：手机号格式校验 → 姓名/手机非空校验 → 默认地址管理（首个自动默认/设新默认取消旧默认） → 数量上限20条
// 接收值：address - 待创建的地址对象指针（UserId由调用方注入）
// 返回值：addressId - 新建地址ID, error - 6002手机号格式/6001姓名手机非空/6003数量上限
func (as *AddressService) CreateAddress(address *model.UserAddress) (addressId int64, err error) {
	// 手机号格式校验（正则：^1[3-9]\d{9}$）
	if !ValidatePhone(address.ReceiverPhone) {
		return 0, model.PhoneMalformed
	}

	// 开启事务
	err = as.DB.Transaction(func(tx *gorm.DB) error {
		// 姓名和手机号非空校验
		if address.ReceiverName == "" || address.ReceiverPhone == "" {
			return model.ReceiverNotNull
		}

		addressTx := as.AddressRepo.WithTx(tx)

		// 默认地址管理：无默认时自动设为默认，有默认时若新地址指定默认则取消旧默认
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
		// 数量上限检查：有效地址不超过20条
		addressList, err := addressTx.GetAddressList(address.UserId)
		if err != nil && !errors.Is(err, model.AddressListIsNull) {
			return err
		}
		if len(addressList) >= 20 {
			return model.AddressNumsIsFull
		}
		// 创建地址
		addressId, err = addressTx.CreateAddress(address)
		if err != nil {
			return err
		}
		return nil
	})
	return addressId, err
}

// GetAddressList 获取用户地址列表
// 功能：查询当前用户所有有效地址（排除已删除），默认地址排在最前，按更新时间降序
// 接收值：userId - 当前用户ID
// 返回值：[]*response.AddressResp - 地址响应列表, error - 6004地址不存在
func (as *AddressService) GetAddressList(userId int64) ([]*response.AddressResp, error) {
	// 调用数据层获取有效地址列表
	addrs, err := as.AddressRepo.GetAddressList(userId)
	if err != nil && !errors.Is(err, model.AddressListIsNull) {
		return nil, err
	}
	// 将数据层模型转换为响应结构体
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

// GetAddress 查询单个地址详情
// 功能：根据地址ID查询详情，并校验地址归属当前用户（防止横向越权）
// 接收值：userId - 当前用户ID, addressId - 地址ID
// 返回值：*response.AddressResp - 地址详情, error - 6004地址不存在/6005无权访问
func (as *AddressService) GetAddress(userId, addressId int64) (*response.AddressResp, error) {
	// 调用数据层查询地址
	address, err := as.AddressRepo.GetAddress(addressId)
	if err != nil {
		return nil, err
	}
	// 归属校验：防止横向越权
	if userId != address.UserId {
		return nil, model.UserNotSetAddress
	}
	// 转换响应结构体
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

// UpdateAddress 更新收货地址
// 功能：在事务内完成归属校验、部分字段覆盖更新、默认地址管理（设新默认时取消旧默认）
// 接收值：userId - 当前用户ID, addressId - 地址ID, updateAddress - 需要更新的字段（map只覆盖传入字段）
// 返回值：*response.AddressResp - 更新后的完整地址, error - 6004地址不存在/6005无权访问
func (as *AddressService) UpdateAddress(userId, addressId int64, updateAddress map[string]interface{}) (*response.AddressResp, error) {
	// 开启事务
	err := as.DB.Transaction(func(tx *gorm.DB) error {
		addressTx := as.AddressRepo.WithTx(tx)
		// 查询原地址并校验归属
		oldAddress, err := addressTx.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != oldAddress.UserId {
			return model.UserNotSetAddress
		}
		// 用原地址初始化，使用mapstructure将请求字段覆盖到新地址对象（仅覆盖传入字段）
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
		// 默认地址管理：若从非默认变为默认，取消旧默认
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
		// 保存更新
		err = addressTx.UpdateAddress(&newAddress)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	// 返回更新后的完整地址
	return as.GetAddress(userId, addressId)
}

// DeleteAddress 删除收货地址（软删除）
// 功能：在事务内校验归属后执行软删除。若删除的是默认地址，自动将最近更新的有效地址设为默认
// 接收值：userId - 当前用户ID, addressId - 地址ID
// 返回值：error - 6004地址不存在/6005无权访问/6006订单引用保护
func (as *AddressService) DeleteAddress(userId, addressId int64) error {
	return as.DB.Transaction(func(tx *gorm.DB) error {
		addressTx := as.AddressRepo.WithTx(tx)
		// 查询地址并校验归属
		address, err := addressTx.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != address.UserId {
			return model.UserNotSetAddress
		}
		// 执行软删除
		err = addressTx.DeleteAddress(addressId)
		if err != nil {
			return err
		}
		// 若删除的是默认地址，将最近更新的有效地址设为新的默认
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

// SetDefaultAddress 设为默认地址
// 功能：在事务内校验归属，清除当前用户的旧默认地址后设置新默认（保证唯一默认约束）
// 接收值：userId - 当前用户ID, addressId - 地址ID
// 返回值：error - 6004地址不存在/6005无权访问
func (as *AddressService) SetDefaultAddress(userId, addressId int64) error {
	return as.DB.Transaction(func(tx *gorm.DB) error {
		// 查询地址并校验归属
		address, err := as.AddressRepo.GetAddress(addressId)
		if err != nil {
			return err
		}
		if userId != address.UserId {
			return model.UserNotSetAddress
		}
		// 事务内操作
		addressTx := as.AddressRepo.WithTx(tx)
		// 查询当前用户的旧默认地址
		defaultAddr, err := addressTx.GetDefaultAddress(userId)
		if err != nil && !errors.Is(err, model.AddressNotExist) {
			return err
		}
		// 取消旧默认
		if defaultAddr != nil {
			err = addressTx.SetCommonAddress(defaultAddr.AddressId)
			if err != nil {
				return err
			}
		}
		// 设置新默认
		err = addressTx.SetDefaultAddress(addressId)
		if err != nil {
			return err
		}
		return nil
	})
}
