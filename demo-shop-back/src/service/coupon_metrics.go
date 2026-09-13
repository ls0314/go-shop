package service

import (
	"errors"
	"time"

	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
)

// ReceiveCoupon 领券入口(带业务指标埋点,DS-A-22)。
// 埋点放在包装层而非散落在原函数的各 return 点:
// 原函数的售罄/超限错误可能来自闸门也可能来自 DB 双防线,业务语义相同,
// 在入口统一按 errors.Is 分类,埋点逻辑与业务逻辑完全解耦
func (c *CouponService) ReceiveCoupon(userId, templateId int64) (*response.UserReceiveCouponResp, error) {
	start := time.Now()
	resp, err := c.receiveCoupon(userId, templateId)

	result := "error"
	switch {
	case err == nil:
		result = "success"
	case errors.Is(err, model.ErrCouponSoldOut):
		result = "sold_out"
	case errors.Is(err, model.ErrCouponLimitExceeded):
		result = "limit_exceeded"
	}
	metrics.CouponReceiveTotal.WithLabelValues(result).Inc()

	// path 标签区分「闸门启用」与「纯 DB」两种链路:
	// 闸门生效时 gate 路径样本暴涨而 db_only 几乎无样本——系统反而最健康
	path := "db_only"
	if c.cache != nil {
		path = "gate"
	}
	metrics.CouponReceiveDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())

	return resp, err
}
