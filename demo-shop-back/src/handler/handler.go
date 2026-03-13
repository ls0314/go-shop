package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {

	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ShouldBindJSON 错误: %v", err) // 👈 打印具体错误！
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误"})
		return
	}

	err := service.Register(req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})

}

func UserInfo(c *gin.Context) {

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	phone, _ := c.Get("phone")
	email, _ := c.Get("email")

	c.JSON(200, gin.H{
		"user_id":  userID,
		"username": username,
		"phone":    phone,
		"email":    email,
	})
}

func LoginHandler(c *gin.Context) {

	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ShouldBindJSON 错误: %v", err) // 👈 打印具体错误！
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误"})
		return
	}

	ip := c.ClientIP()
	device := c.Request.UserAgent()

	resp, err := service.Login(req, ip, device)

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, resp)
}
