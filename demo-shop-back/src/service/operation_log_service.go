package service

import (
	"demo-shop-back/src/model/requset"
	"demo-shop-back/src/model/response"
	"demo-shop-back/src/repository"
)

type OperationLogService struct {
	OperationLogRepo *repository.OperationLogRepo
}

func NewOperationLogService() *OperationLogService {
	return &OperationLogService{
		OperationLogRepo: repository.NewOperationLogRepo(),
	}
}

func (o *OperationLogService) GetOperationLogList(req requset.GetOperationLogList) (*response.GetOperationLogListResp, error) {
	// 防止参数越界
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	operationLogList, total, err := o.OperationLogRepo.GetOperationLogList(req)
	if err != nil {
		return nil, err
	}

	resp := &response.GetOperationLogListResp{
		List:     operationLogList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	return resp, nil
}
