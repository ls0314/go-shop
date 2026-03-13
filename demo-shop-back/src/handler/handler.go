package handler

import (
	"bytes"
	"demo-shop-back/src/model"
	"demo-shop-back/src/service"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	log.Println("==================================")
	log.Println("=== REGISTER HANDLER INVOKED ====")
	log.Println("==================================")

	var req model.RegisterRequest

	body, _ := c.GetRawData()
	log.Println("请求body:", string(body))
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

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
