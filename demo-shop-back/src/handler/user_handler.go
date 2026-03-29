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
		//c.JSON(http.StatusBadRequest, gin.H{
		//	"code":    200,
		//	"message": "请求参数错误",
		//	"data":    nil,
		//})
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	err := service.Register(req)

	if err != nil {
		//c.JSON(http.StatusUnauthorized, gin.H{
		//	"code":    200,
		//	"message": err.Error(),
		//	"data":    nil,
		//})
		utils.Error(c, 500, model.StatusInternalServerError+err.Error())
		return
	}
	//c.JSON(http.StatusOK, gin.H{
	//	"code":    200,
	//	"message": "注册成功",
	//	"data":    nil,
	//})
	utils.Success(c, nil)

}

func GetUserInfo(c *gin.Context) {

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	//c.JSON(http.StatusOK, gin.H{
	//	"code":    200,
	//	"message": "请求成功",
	//	"data": gin.H{"user_id": userID,
	//		"username": username},
	//})
	utils.Success(c, gin.H{"user_id": userID,
		"username": username})
}

func LoginHandler(c *gin.Context) {

	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{
		//	"code":    200,
		//	"message": "请求参数错误",
		//	"data":    nil})
		utils.Fail(c, 400, model.StatusBadRequest+err.Error())
		return
	}

	ip := c.ClientIP()
	device := c.Request.UserAgent()

	resp, err := service.Login(req, ip, device)

	if err != nil {

		//c.JSON(http.StatusUnauthorized, gin.H{
		//	"code":    200,
		//	"message": err.Error(),
		//	"data":    nil,
		//})
		utils.Error(c, 500, model.StatusInternalServerError+err.Error())
		return
	}

	//c.JSON(http.StatusOK, gin.H{
	//	"code":    200,
	//	"message": "登录成功",
	//	"data":    resp,
	//})
	utils.Success(c, resp)
}

func RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		//c.JSON(http.StatusBadRequest, gin.H{
		//	"code":    400,
		//	"message": "参数错误",
		//	"data":    nil,
		//})
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	jwtService := middleware.GetJWTService()

	claims, err := jwtService.ParseToken(req.RefreshToken)
	if err != nil {
		//c.JSON(http.StatusUnauthorized, gin.H{"message": "RefreshToken无效/已过期"})
		utils.Fail(c, 400, "RefreshToken无效/已过期")
		return
	}

	if claims.TokenType != "refresh" {
		//c.JSON(http.StatusUnauthorized, gin.H{"message": "Token类型错误"})
		utils.Fail(c, 400, "Token类型错误")
		return
	}

	newAccessToken, err := jwtService.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		//c.JSON(http.StatusInternalServerError, gin.H{"message": "生成Token失败"})
		utils.Fail(c, 400, "生成Token失败")
		return
	}

	//c.JSON(http.StatusOK, gin.H{
	//	"code":    200,
	//	"message": "刷新成功",
	//	"data": gin.H{
	//		"access_token": newAccessToken,
	//	},
	//})
	utils.Success(c, gin.H{
		"access_token": newAccessToken,
	})
}
