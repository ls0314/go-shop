package service

import (
	"errors"
	"time"

	"demo-shop-back/src/infra/metrics"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
)

// ReceiveCoupon 领券入口(带业务指标埋点,DS-A-22)。
//
// 埋点放在包装层而非散落在各 return 点:售罄/超限这两个业务失败
// 在服务端可能来自闸门也可能来自 DB 双防线,业务语义相同,
// 在入口统一按 errors.Is 分类,埋点逻辑与业务逻辑完全解耦。
//
// **与迁移前的一处变化**:原来还有一条 path 标签区分「闸门启用(gate)」与
// 「纯 DB(db_only)」链路 —— 那个判断读的是本地 c.cache 是否注入。
// 券域迁走后闸门在 marketing-service 进程内,单体观察不到它是否生效,
// 强行保留只会得到一个恒为 db_only 的假标签,故移除该标签的区分度,
// 只保留结果维度。链路归属请查 marketing-service 侧的指标。
func (c *CouponService) ReceiveCoupon(userId, templateId int64) (*response.UserReceiveCouponResp, error) {
	start := time.Now()
	resp, err := c.receiveCouponRPC(userId, templateId)

	result := "error"
	switch {
	case errMsgIs(err, model.ErrCouponSoldOut):
		result = "sold_out"
	case errMsgIs(err, model.ErrCouponLimitExceeded):
		result = "limit_exceeded"
	case err == nil && resp != nil:
		result = "success"
	}
	metrics.CouponReceiveTotal.WithLabelValues(result).Inc()
	metrics.CouponReceiveDuration.WithLabelValues("rpc").Observe(time.Since(start).Seconds())

	return resp, err
}

// errMsgIs 按错误文案判等。
//
// 为什么不用 errors.Is:调用方拿到的是 RPC 客户端用 error_msg 新建的错误
// (见 couponclient.RestoreError),与本地哨兵变量不是同一个实例。
// 文案是跨服务契约,这里按文案比对;errors.Is 仍然先试一次,
// 以便本地也能直接返回哨兵错误(如客户端未建连)。
func errMsgIs(err error, target error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, target) || err.Error() == target.Error()
}
