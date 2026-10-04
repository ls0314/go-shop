package req

import (
	"errors"
	"time"

	"demo-shop/services/trade/internal/model"
)

// IdempotentKeyMinLen/IdempotentKeyMaxLen 幂等键长度约束。
const (
	IdempotentKeyMinLen = 16
	IdempotentKeyMaxLen = 40
)

// 下单相关的业务校验错误。
var (
	// ErrIdempotentKeyRequired 未传幂等键。
	ErrIdempotentKeyRequired = errors.New("缺少幂等键")
	// ErrIdempotentKeyInvalid 幂等键长度不合法
	ErrIdempotentKeyInvalid = errors.New("幂等键格式不合法")
	// ErrAddressSnapshotRequired 未传收货地址快照
	ErrAddressSnapshotRequired = errors.New("缺少收货地址")
)

// Validate 校验入参形状。
func (r *CreateOrderReq) Validate() error {
	n := len(r.IdempotentKey)
	if n == 0 {
		return ErrIdempotentKeyRequired
	}
	if n < IdempotentKeyMinLen || n > IdempotentKeyMaxLen {
		return ErrIdempotentKeyInvalid
	}
	if r.AddressSnapshot.ReceiverName == "" || r.AddressSnapshot.ReceiverPhone == "" {
		return ErrAddressSnapshotRequired
	}
	return nil
}

// CreateOrderReq 下单入参(领域层)。
type CreateOrderReq struct {
	UserId   int64
	UserName string
	// AddressSnapshot 收货地址快照,由调用方在结算页选定地址后传入。
	AddressSnapshot model.AddressSnap
	// IdempotentKey 幂等键,由调用方(前端)生成、全局唯一、**重试时不变**。
	IdempotentKey string
	BuyerRemark   string
	// UserCouponId 为 0 表示不使用优惠券
	UserCouponId int64
}

// ListUserOrdersReq 用户端订单列表查询
type ListUserOrdersReq struct {
	Page        int
	PageSize    int
	OrderStatus string
}

// CancelOrderReq 取消订单。
type CancelOrderReq struct {
	OrderId  int64
	UserId   int64
	Operator string
}

// ShipOrderReq 管理端发货
type ShipOrderReq struct {
	OrderId        int64
	UserName       string
	ExpressCompany string
	TrackingNo     string
}

// ListOrdersReq 管理端订单列表查询
type ListOrdersReq struct {
	Page        int
	PageSize    int
	OrderStatus string
	OrderNo     string
	StartTime   *time.Time
	EndTime     *time.Time
}
