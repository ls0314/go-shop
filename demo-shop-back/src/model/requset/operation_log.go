package requset

import "time"

type GetOperationLogList struct {
	Page          int        `json:"page"`
	PageSize      int        `json:"page_size"`
	Module        string     `json:"module"`
	RequestMethod string     `json:"request_method"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
}
