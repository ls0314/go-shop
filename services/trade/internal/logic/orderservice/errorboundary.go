package orderservicelogic

import (
	"errors"

	"demo-shop/services/trade/internal/dto/req"
	"demo-shop/services/trade/internal/model"

	"demo-shop/services/trade/internal/infra/couponclient"
	"demo-shop/services/trade/internal/infra/productclient"
)

// ============================================================
// 错误边界:领域错误 ↔ gRPC 的两种失败形态
// ============================================================
var bizErrors = []error{
	// 入参形状
	req.ErrIdempotentKeyRequired,
	req.ErrIdempotentKeyInvalid,
	req.ErrAddressSnapshotRequired,

	// 购物车
	model.ErrCartNoSettlementItems,
	model.CartItemNotExist,
	model.NotAuthority,
	model.ErrSpuDisabled,
	model.CartItemMax,
	model.QuantityMax,

	// 订单状态机
	model.ErrOrderNotExist,
	model.ErrOrderNoPermission,
	model.ErrOrderCannotCancel,
	model.ErrOrderAlreadyCancelled,
	model.ErrOrderCannotShip,
	model.ErrOrderCannotConfirm,
	model.ErrOrderCannotPay,

	// 库存与商品(经 product RPC 还原回来的领域错误)
	model.ErrSkuNotExist,
	model.ErrStockNotEnough,
	model.ErrSkuDisabled,
	model.ErrLockStockNotEnough,

	// 券(经 marketing RPC 还原回来的领域错误)
	model.ErrCouponNotExistOrUsed,
	model.ErrCouponThresholdNotMet,
	model.ErrCouponSoldOut,
	model.ErrCouponLimitExceeded,
	model.ErrCouponTemplateNotExist,
	model.ErrUseCouponNoNoPermission,
	model.ErrCannotCancelCoupon,

	// 支付
	model.ErrPayNoPermission,
	model.ErrPayRecordExisting,
	model.ErrPayRecordNoExist,
	model.ErrPayRecordNoNoPermission,
	model.ErrPayStatusMisTake,
	model.ErrPayAmountMisTake,
}

// isBizError 是否是可预期的业务失败
func isBizError(err error) bool {
	if err == nil {
		return false
	}
	for _, target := range bizErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// bizErrMsg 把错误翻成 error_msg。
//
// 业务失败 → 文案;基础设施失败 → 空串(调用方据 gRPC error 判)。
func bizErrMsg(err error) string {
	if isBizError(err) {
		return err.Error()
	}
	return ""
}

// couponErrIs / productErrIs 判定下游客户端的"服务不可用"哨兵。
//
// 单独抽出来是因为它俩的文案是**契约**(见各自 client 的 ErrUnavailable):
// 上游据此回 503,接线冒烟测试据此跳过该路由。
func isDownstreamUnavailable(err error) bool {
	return errors.Is(err, productclient.ErrUnavailable) ||
		errors.Is(err, couponclient.ErrUnavailable)
}
