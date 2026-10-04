package model

import (
	"errors"
)

// ---- 订单域错误码 ----

var (
	// ErrSkuNotExist SKU不存在(7001)
	ErrSkuNotExist = errors.New("SKU不存在")
	// ErrStockNotEnough 库存不足(7004)
	ErrStockNotEnough = errors.New("库存不足")
	// ErrSkuDisabled SKU已禁用或已删除(7005)
	ErrSkuDisabled = errors.New("SKU已禁用或已删除")
	// ErrLockStockNotEnough 锁定库存不足(7006)
	ErrLockStockNotEnough = errors.New("锁定库存不足")
)

// ---- 购物车模块 ----
var (
	// ErrSpuDisabled 商品已下架,加购被拒
	ErrSpuDisabled = errors.New("SKU已禁用或已删除")
	// CartItemMax 购物车行数上限
	CartItemMax = errors.New("购物车数量已达上限（100条）")
	// QuantityMax 单 SKU 购买数量上限
	QuantityMax = errors.New("购买数量超出限制（单SKU购买范围为0~999）")
	// CartItemNotExist 购物车项不存在
	CartItemNotExist = errors.New("购物车项不存在")
	// NotAuthority 越权操作他人购物车
	NotAuthority = errors.New("无权操作该购物车项")
)

// ---- 订单模块 ----
var (
	// ErrAddressNotExist 收货地址不存在
	ErrAddressNotExist = errors.New("收货地址不存在")
	// ErrCartNoSettlementItems 无可结算商品(未选中任何行)
	ErrCartNoSettlementItems = errors.New("购物车无可结算商品")
	// ErrOrderNotExist 订单不存在
	ErrOrderNotExist = errors.New("订单不存在")
	// ErrOrderNoPermission 越权查看/操作他人订单
	ErrOrderNoPermission = errors.New("无权查看/操作该订单")
	// ErrOrderCannotCancel 仅待支付可取消
	ErrOrderCannotCancel = errors.New("订单状态不允许取消（仅待支付可取消）")
	// ErrOrderAlreadyCancelled 已取消或已支付,无需重复取消
	ErrOrderAlreadyCancelled = errors.New("订单已被处理（已取消或已支付），无需重复取消")
	// ErrOrderCannotShip 仅已支付可发货
	ErrOrderCannotShip = errors.New("订单状态不允许发货（仅已支付可发货）")
	// ErrOrderCannotConfirm 仅已发货可确认收货
	ErrOrderCannotConfirm = errors.New("订单状态不允许确认收货（仅已发货可确认）")
)

// ---- 支付模块 ----
var (
	// ErrPayNoPermission 订单不属于当前用户
	ErrPayNoPermission = errors.New("订单不属于当前用户")
	// ErrOrderCannotPay 仅待支付可支付
	ErrOrderCannotPay = errors.New("订单状态不允许支付（仅待支付可支付）")
	// ErrPayRecordExisting 已有进行中的支付记录
	ErrPayRecordExisting = errors.New("已有进行中的支付记录")
	// ErrPayRecordNoExist 支付记录不存在
	ErrPayRecordNoExist = errors.New("支付记录不存在")
	// ErrPayRecordNoNoPermission 越权查看他人支付记录
	ErrPayRecordNoNoPermission = errors.New("无权查看该支付记录")
	// ErrPayStatusMisTake 非 pending,回调幂等
	ErrPayStatusMisTake = errors.New("支付状态不正确（非pending，回调幂等）")
	// ErrPayAmountMisTake 支付金额与订单金额不匹配
	ErrPayAmountMisTake = errors.New("支付金额与订单金额不匹配")
	// ErrCannotDeductAfterPay 支付已成功但库存扣减失败。
	//
	// 这是**向前补偿**的信号,不是普通业务失败:支付不可逆、不能退钱,
	// 故返回错误让渠道重试回调(重放会被 pay_status 的幂等挡住,不会重复扣款),
	// 渠道不再重试时由对账任务收敛。
	// 单独一条文案是为了让日志与告警能一眼认出这类"钱收了、货没出"的记录。
	ErrCannotDeductAfterPay = errors.New("支付已成功但库存扣减失败")
)

// ---- 券域 ----
var (
	// ErrCouponNotExistOrUsed 券不存在或已使用
	ErrCouponNotExistOrUsed = errors.New("优惠券不存在或已使用")
	// ErrCouponThresholdNotMet 未达使用门槛
	ErrCouponThresholdNotMet = errors.New("优惠券不满足使用门槛")
	// ErrCouponSoldOut 券已领完
	ErrCouponSoldOut = errors.New("优惠券已领完")
	// ErrCouponLimitExceeded 已达领取上限
	ErrCouponLimitExceeded = errors.New("已达领取上限")
	// ErrCouponTemplateNotExist 券模板不存在
	ErrCouponTemplateNotExist = errors.New("优惠券模板不存在")
	// ErrUseCouponNoNoPermission 越权用他人券
	ErrUseCouponNoNoPermission = errors.New("无权使用该优惠券")
	// ErrCannotCancelCoupon 取消订单归还券失败
	ErrCannotCancelCoupon = errors.New("返还优惠券失败")
)

// ---- 依赖不可用(调用方据此回 503) ----
var (
	// ErrProductUnavailable product-service 未建连
	ErrProductUnavailable = errors.New("product-service 不可用")
	// ErrMarketingUnavailable marketing-service 未建连
	ErrMarketingUnavailable = errors.New("marketing-service 不可用")
)
