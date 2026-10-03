package utils

import (
	"errors"
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
	_ = c.Error(errors.New(message))
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func Error(c *gin.Context, code int, message string) {
	_ = c.Error(errors.New(message))
	c.JSON(http.StatusInternalServerError, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Unavailable 下游微服务未就绪(未建连/连不上)时的统一响应:503 + 约定文案。
//
// 为什么与 Error(500) 分开:
//   - 语义不同。500 是"本服务出错了",503 是"依赖的服务暂时不可用";
//   - 可判据不同。运维与测试需要能一眼区分"接线断了"与"依赖没起来",
//     否则接线冒烟测试只能把两者一起报错,产生假红。
//
// 契约:消息以 " <服务名> 不可用" 结尾(见 tests/wiring_smoke_test.go 的
// downstreamUnavailable),改文案要同步改那里。
func Unavailable(c *gin.Context, message string) {
	_ = c.Error(errors.New(message))
	c.JSON(http.StatusServiceUnavailable, Response{
		Code:    http.StatusServiceUnavailable,
		Message: message,
		Data:    nil,
	})
}
