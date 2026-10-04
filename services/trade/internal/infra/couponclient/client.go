package couponclient

import (
	"context"
	"errors"
	"time"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/trade/internal/model"
)

// ErrUnavailable 客户端未建连时的统一错误
var ErrUnavailable = errors.New("marketing-service 不可用")

// callTimeout 单次券 RPC 的超时。
const callTimeout = 5 * time.Second

// Client 券域 RPC 客户端,实现 service.CouponRPC。
type Client struct {
	c v1_marketingv1.CouponServiceClient
}

func NewClient(c v1_marketingv1.CouponServiceClient) *Client {
	return &Client{c: c}
}

// UseCoupon 核销券,返回用券后的实付金额。
func (c *Client) UseCoupon(userCouponId int64, idempotencyKey, orderNo string, userId int64, orderAmount float64) (float64, error) {
	if c == nil || c.c == nil {
		return 0, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.c.UseCoupon(ctx, &v1_marketingv1.UseCouponReq{
		UserCouponId:   userCouponId,
		IdempotencyKey: idempotencyKey,
		OrderNo:        orderNo,
		UserId:         userId,
		OrderAmount:    orderAmount,
	})
	if err != nil {
		return 0, err
	}
	if err := RestoreError(resp.ErrorMsg); err != nil {
		return 0, err
	}
	return resp.PayAmount, nil
}

// ReturnCoupon 按幂等键归还券(补偿)。
func (c *Client) ReturnCoupon(idempotencyKey string, userId int64) error {
	if c == nil || c.c == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.c.ReturnCoupon(ctx, &v1_marketingv1.ReturnCouponReq{
		IdempotencyKey: idempotencyKey,
		UserId:         userId,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// RestoreError 把服务端 error_msg 还原成本地哨兵错误(能还原时)或普通错误。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if err, ok := serverErrMap[errorMsg]; ok {
		return err
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// serverErrMap 服务端文案 → 本地哨兵错误。
var serverErrMap = map[string]error{
	"优惠券不存在或已使用": model.ErrCouponNotExistOrUsed,
	"优惠券不满足使用门槛": model.ErrCouponThresholdNotMet,
	"优惠券已领完":     model.ErrCouponSoldOut,
	"已达领取上限":     model.ErrCouponLimitExceeded,
	"优惠券模板不存在":   model.ErrCouponTemplateNotExist,
	"无权使用该优惠券":   model.ErrUseCouponNoNoPermission,
	"返还优惠券失败":    model.ErrCannotCancelCoupon,
}
