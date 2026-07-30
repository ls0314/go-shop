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
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	PayStatus string     `json:"pay_status"`
	PayMethod string     `json:"pay_method"`
	OrderNo   string     `json:"order_no"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}
