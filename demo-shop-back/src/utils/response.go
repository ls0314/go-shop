package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Success",
		Data:    data,
	})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
