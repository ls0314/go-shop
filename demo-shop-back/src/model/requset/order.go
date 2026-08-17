package requset

import "time"

type CreatOrderReq struct {
	AddressId     int64  `json:"address_id"`
	IdempotentKey string `json:"idempotent_key"`
	BuyerRemark   string `json:"buyer_remark"`
	UserCouponId  int64  `json:"user_coupon_id"`
}

type UserGetOrderListReq struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	OrderStatus string `form:"order_status"`
}

type GetOrderListReq struct {
	Page        int        `form:"page"`
	PageSize    int        `form:"page_size"`
	OrderStatus string     `form:"order_status"`
	OrderNo     string     `form:"order_no"`
	StartTime   *time.Time `form:"start_time"`
	EndTime     *time.Time `form:"end_time"`
}

type OrderShipReq struct {
	ExpressCompany string `json:"express_company"`
	TrackingNO     string `json:"tracking_no"`
}
