package resp

import (
	"time"

	"demo-shop/services/trade/internal/model"
)

// CreateOrderResp 下单结果(聚合输出)。
type CreateOrderResp struct {
	OrderId     int64
	OrderNo     string
	TotalAmount float64
	PayAmount   float64
	OrderStatus string
	// PayExpireAt 支付截止时间。
	PayExpireAt time.Time
	CreatedAt   time.Time
}

// OrderDetailView 订单详情(订单 + 明细 + 日志)。
type OrderDetailView struct {
	Order   *model.UserOrder
	Details []*model.UserOrderDetail
	Logs    []*model.UserOrderLog
}

// CancelOrderResp 取消结果。
type CancelOrderResp struct {
	Order *model.UserOrder
	// Compensated 库存释放与券归还是否都已成功。
	Compensated bool
}

// OrderPage 订单分页结果(用户端与管理端共用)
type OrderPage struct {
	Items    []*model.UserOrder
	Total    int64
	Page     int
	PageSize int
}

// ShipOrderResp 发货结果
type ShipOrderResp struct {
	Order          *model.UserOrder
	ExpressCompany string
	TrackingNo     string
}
