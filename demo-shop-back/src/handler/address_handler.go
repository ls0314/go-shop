package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	AddressService *service.AddressService
}

func NewAddressHandler() *AddressHandler {
	return &AddressHandler{
		AddressService: service.NewAddressService(),
	}
}

func GetUserId(c *gin.Context) (int64, error) {
	userIdVal, exist := c.Get("user_id")
	if !exist {
		return 0, model.UserNotLogin
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		return 0, model.UserNotLogin
	}
	return userId, nil
}

func (ah *AddressHandler) CreateAddress(c *gin.Context) {
	var address model.UserAddress
	if err := c.ShouldBind(&address); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	address.UserId = userId
	addressId, err := ah.AddressService.CreateAddress(&address)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, addressId)
}

func (ah *AddressHandler) GetAddressList(c *gin.Context) {
	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	addressList, err := ah.AddressService.GetAddressList(userId)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, addressList)
}

func (ah *AddressHandler) GetAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	address, err := ah.AddressService.GetAddress(userId, id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, address)
}

func (ah *AddressHandler) UpdateAddress(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	var updateAddress map[string]interface{}

	if err := c.ShouldBind(&updateAddress); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	address, err := ah.AddressService.UpdateAddress(userId, id, updateAddress)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, address)
}

func (ah *AddressHandler) DeleteAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	err = ah.AddressService.DeleteAddress(userId, id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

func (ah *AddressHandler) SetDefaultAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusIdNotExist+err.Error())
		return
	}

	userId, err := GetUserId(c)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	err = ah.AddressService.SetDefaultAddress(userId, id)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
