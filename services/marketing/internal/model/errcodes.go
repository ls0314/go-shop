package model

import "errors"

// 券域错误码。编号与文案与单体 model/error_info.go **逐字一致** ——
// RPC 的 error_msg 传的是文案,调用方(单体/BFF)按文案还原成本地错误变量,
// 因此文案必须一字不差,errors.Is 判等才成立。改文案 = 改跨服务契约。
var (
	// ErrCouponTemplateNotExist 模板不存在(11001)
	ErrCouponTemplateNotExist = errors.New("优惠券模板不存在")
	// ErrCouponSoldOut 已领完(11002)
	ErrCouponSoldOut = errors.New("优惠券已领完")
	// ErrCouponLimitExceeded 已达每人限领上限(11003)
	ErrCouponLimitExceeded = errors.New("已达领取上限")
	// ErrCouponNotExistOrUsed 券不存在/已使用/已过期(11004)
	ErrCouponNotExistOrUsed = errors.New("优惠券不存在或已使用")
	// ErrCouponThresholdNotMet 未达使用门槛(11005)
	ErrCouponThresholdNotMet = errors.New("优惠券不满足使用门槛")
	// ErrCouponExpired 券已过期(11006)。
	// 当前实现用查询条件 expire_at > now 表达"未过期",不显式抛此错;
	// 保留定义以与单体的错误码表完整对齐(调用方可能按此判定)。
	ErrCouponExpired = errors.New("优惠券已过期")
	// ErrCouponParamInvalid 创建参数非法(11007)
	ErrCouponParamInvalid = errors.New("参数非法（类型/金额/数量校验失败）")
	// ErrCouponValidityInvalid 有效期两种模式均未配置或同时配置(11008)
	ErrCouponValidityInvalid = errors.New("有效期配置非法（两种有效期方式均未配置或同时配置）")
	// ErrCouponOrderAmountInvalid 订单金额参数非法(11009)
	ErrCouponOrderAmountInvalid = errors.New("订单金额参数非法（小于 0）")
	// ErrUseCouponNoPermission 券不属于该用户。
	// 单体原文案是"无权使用该优惠卷"(卷/券为笔误),此处改正 ——
	// 该错误只可能由本服务返回,调用方按文案还原时需同步。
	ErrUseCouponNoPermission = errors.New("无权使用该优惠券")
	// ErrCannotCancelCoupon 取消订单归还券失败。
	// 单体原文案"返还优惠卷失败",同上改正。
	ErrCannotCancelCoupon = errors.New("返还优惠券失败")
	// ErrIdempotencyKeyRequired 缺少幂等键。
	//
	// 核销与归还都靠它做判重与补偿反查,缺了就不能放行 ——
	// 宁可显式失败,也不要"静默无幂等"地核销一张券
	ErrIdempotencyKeyRequired = errors.New("缺少幂等键")
)
