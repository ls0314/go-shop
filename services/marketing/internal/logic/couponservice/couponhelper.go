package couponservicelogic

import (
	"errors"
	"time"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/model"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// 券域可预期的业务失败。落在此集合内以 error_msg 返回(gRPC error 为 nil),
// 其余视为基础设施故障,gRPC error 非 nil 由调用方重试。
// 与 user-service 的同类归类同一约定。
var couponBizErrors = []error{
	model.ErrCouponTemplateNotExist,
	model.ErrCouponSoldOut,
	model.ErrCouponLimitExceeded,
	model.ErrCouponNotExistOrUsed,
	model.ErrCouponThresholdNotMet,
	model.ErrCouponExpired,
	model.ErrCouponParamInvalid,
	model.ErrCouponValidityInvalid,
	model.ErrCouponOrderAmountInvalid,
	model.ErrUseCouponNoPermission,
	model.ErrCannotCancelCoupon,
}

func isCouponBizError(err error) bool {
	for _, target := range couponBizErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// 分页上限。管理端与用户端同值 —— 券列表没有管理端要拉全量的场景。
const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 统一分页参数:page<=0 → 1;pageSize<=0 → 10;>100 封顶 100。
// 与 user-service 的 ListScopes 同一口径。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// FieldUpdatesToMap 把 proto 的 repeated FieldUpdate 转成"字段名 → 值"的映射,
// 供 mapstructure 按 json tag 合并。与 user-service 的 converter 同构。
func FieldUpdatesToMap(updates []*v1_marketingv1.FieldUpdate) map[string]interface{} {
	out := make(map[string]interface{}, len(updates))
	for _, u := range updates {
		if u == nil || u.Field == "" {
			continue
		}
		switch v := u.Value.(type) {
		case *v1_marketingv1.FieldUpdate_StringValue:
			out[u.Field] = v.StringValue
		case *v1_marketingv1.FieldUpdate_Int64Value:
			out[u.Field] = v.Int64Value
		case *v1_marketingv1.FieldUpdate_DoubleValue:
			out[u.Field] = v.DoubleValue
		case *v1_marketingv1.FieldUpdate_BoolValue:
			out[u.Field] = v.BoolValue
		}
	}
	return out
}

// decodeUpdates 把 updates 合并进 dst(按 json tag 匹配字段名)。
// 传入空 updates 时不做任何改动,返回 nil 而非报错 ——
// "没传要改的字段"是合法入参,等价于不更新。
func decodeUpdates(dst interface{}, updates []*v1_marketingv1.FieldUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  dst,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(FieldUpdatesToMap(updates))
}

// calcExpireAt 按模板的有效期模式算券的过期时间。
// 相对(usable_days > 0)从领取时刻起算,否则用固定 end_time —— 与单体一致。
func calcExpireAt(tpl *model.CouponTemplate, now time.Time) time.Time {
	if tpl.UsableDays > 0 {
		return now.AddDate(0, 0, int(tpl.UsableDays))
	}
	return tpl.EndTime
}

// validateTemplate 创建参数校验,与单体 CreateCoupon 的校验链逐条对应。
// 返回 nil 表示通过。
func validateTemplate(tpl *model.CouponTemplate) error {
	// 类型必须为两种枚举之一(与 DB 的 ck_coupon_type 对应)
	if tpl.CouponType != model.CouponTypeFullReduction && tpl.CouponType != model.CouponTypeDirectDiscount {
		return model.ErrCouponParamInvalid
	}
	// 优惠力度必须 > 0(满减为金额、直减为折扣率,均不允许 0/负数)
	if tpl.DiscountAmount <= 0 {
		return model.ErrCouponParamInvalid
	}
	// 使用门槛 >= 0(0 表示无门槛券)
	if tpl.ThresholdAmount < 0 {
		return model.ErrCouponParamInvalid
	}
	// 发放总量必须 > 0
	if tpl.TotalCount <= 0 {
		return model.ErrCouponParamInvalid
	}
	// 每人限领不能超过发放总量(否则限领失去意义)
	if tpl.PerUserLimit > tpl.TotalCount {
		return model.ErrCouponParamInvalid
	}
	// 有效期两种模式必须二选一:相对未配 + 固定未配对 → 11008
	if tpl.UsableDays == 0 && (tpl.StartTime.IsZero() || tpl.EndTime.IsZero()) {
		return model.ErrCouponValidityInvalid
	}
	return nil
}

// couponErrText 把 repo 层透传上来的错误翻成对外的 error_msg 文案。
//
// repo 遵循 user-service 的约定直接返回 gorm.ErrRecordNotFound(不替调用方
// 决定语义),由这里翻成券域的业务文案 —— 文案是跨服务契约,调用方按它
// 还原成本地错误变量。
func couponErrText(err error) string {
	if err == gorm.ErrRecordNotFound {
		return model.ErrCouponTemplateNotExist.Error()
	}
	return err.Error()
}
