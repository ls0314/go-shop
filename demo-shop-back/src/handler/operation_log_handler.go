package handler

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/service"
	"demo-shop-back/src/utils"

	"github.com/gin-gonic/gin"
)

type OperationLogHandler struct {
	OperationLogService *service.OperationLogService
}

func NewOperationLogHandler() *OperationLogHandler {
	return &OperationLogHandler{
		OperationLogService: service.NewOperationLogService(),
	}
}

func (o *OperationLogHandler) GetOperationLogList(c *gin.Context) {
	var req requset.GetOperationLogList
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Fail(c, 400, model.StatusBadRequest)
		return
	}

	resp, err := o.OperationLogService.GetOperationLogList(req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
