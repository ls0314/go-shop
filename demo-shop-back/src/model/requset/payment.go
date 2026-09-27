package requset

import "time"

type CreatePaymentReq struct {
	PayMethod string `json:"pay_method"`
}

type GetPaymentReq struct {
	PayNo string `json:"pay_no"`
}

type CallbackPaymentReq struct {
	PayNo   string `json:"pay_no"`
	TradeNo string `json:"trade_no"`
}

type GetPaymentListReq struct {
	Page      int        `form:"page"`
	PageSize  int        `form:"page_size"`
	PayStatus string     `form:"pay_status"`
	PayMethod string     `form:"pay_method"`
	OrderNo   string     `form:"order_no"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
}
