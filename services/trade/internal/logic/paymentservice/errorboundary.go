package paymentservicelogic

import (
	"errors"

	"demo-shop/services/trade/internal/model"

	"demo-shop/services/trade/internal/infra/productclient"
)

// ============================================================
// 错误边界:领域错误 → error_msg
// ============================================================
var bizErrors = []error{
	model.ErrOrderNotExist,
	model.ErrPayNoPermission,
	model.ErrOrderCannotPay,
	model.ErrPayRecordExisting,
	model.ErrPayRecordNoExist,
	model.ErrPayRecordNoNoPermission,
	model.ErrPayStatusMisTake,
	model.ErrPayAmountMisTake,
}

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

// bizErrMsg 业务失败 → 文案;基础设施失败 → 空串
func bizErrMsg(err error) string {
	if isBizError(err) {
		return err.Error()
	}
	return ""
}

// isDownstreamUnavailable 下游未建连(文案是契约)
func isDownstreamUnavailable(err error) bool {
	return errors.Is(err, productclient.ErrUnavailable)
}
