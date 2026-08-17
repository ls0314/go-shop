package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	CouponService *service.CouponService
}

func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		CouponService: service.NewCouponService(),
	}
}

func (ch *CouponHandler) CreateCouponTemplate(c *gin.Context) {
	var req requset.CreateCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := ch.CouponService.CreateCoupon(req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ch *CouponHandler) GetCouponList(c *gin.Context) {
	var req requset.GetCouponListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}
	resp, err := ch.CouponService.GetCouponList(req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ch *CouponHandler) GetUserCouponList(c *gin.Context) {
	var req requset.UserGetCouponListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}
	resp, err := ch.CouponService.UserGetCouponList(userId, req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ch *CouponHandler) ReceiveCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}

	resp, err := ch.CouponService.ReceiveCoupon(userId, id)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func (ch *CouponHandler) GetAvailableCouponList(c *gin.Context) {
	var req requset.GetAvailableCouponReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	userId, _, err := GetUserInfoByContext(c)
	if err != nil {
		utils.Fail(c, 500, model.UserInfoError.Error())
		return
	}

	resp, err := ch.CouponService.GetAvailableCouponList(userId, req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
