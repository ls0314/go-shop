package response

import "demo-shop-back/src/model"

type GetOperationLogListResp struct {
	List     []*model.OperationLog `json:"list"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}
