package handler

import (
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {

	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	err := service.Register(req)

	if err != nil {
		utils.Error(c, 500, model.StatusInternalServerError+err.Error())
		return
	}
	utils.Success(c, nil)

}

func GetUserInfo(c *gin.Context) {

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	utils.Success(c, gin.H{"user_id": userID,
		"username": username})
}

func LoginHandler(c *gin.Context) {

	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	ip := c.ClientIP()
	device := c.Request.UserAgent()

	resp, err := service.Login(req, ip, device)

	if err != nil {

		utils.Error(c, 500, model.StatusInternalServerError+err.Error())
		return
	}

	utils.Success(c, resp)
}

func RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	jwtService := middleware.GetJWTService()

	claims, err := jwtService.ParseToken(req.RefreshToken)
	if err != nil {
		utils.Fail(c, 400, "RefreshToken无效/已过期")
		return
	}

	if claims.TokenType != "refresh" {
		utils.Fail(c, 400, "Token类型错误")
		return
	}

	newAccessToken, err := jwtService.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		utils.Fail(c, 400, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"access_token": newAccessToken,
	})
}
