package addressservicelogic

import (
	"errors"
	"regexp"
	"time"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// ============================================================
// 地址域的共用辅助
// ============================================================

// phonePattern 手机号格式。与单体的 `^1[3-9]\d{9}$` 逐字一致 ——
// 表上还有一条同规则的 CHECK 约束(ck_receiver_phone),两处必须是同一套规则;
// 拆开维护必然漂移,而漂移的表现是"应用放过了、DB 拒绝"的 500 而不是业务错误。
var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// validateAddressFields 校验收货人与手机号。
//
// 顺序照搬单体:**先查手机号格式,再查非空**。
// 看起来反了(空手机号会先被判"格式错误"而不是"不能为空"),但文案是
// 前端展示的契约,换顺序会改变用户看到的提示 —— 迁出时保持不动。
func validateAddressFields(name, phone string) error {
	if !phonePattern.MatchString(phone) {
		return model.PhoneMalformed
	}
	if name == "" || phone == "" {
		return model.ReceiverNotNull
	}
	return nil
}

// loadOwnedAddress 取地址并校验归属。
//
// 归属校验放在**表的所有权方**(这里),而不是让调用方传一个"已校验"的标记:
// 调用方可能传错 userId,不能替它兜底。
//
// 查不到与不属于当前用户返回**两个不同的错误**:前者是 404 语义,
// 后者是 403 语义。单体这两个都用 UserNotSetAddress 表达(措辞偏向前者),
// 迁出时保持不动以免改变前端提示,但函数内部把两者分开,便于日志与
// 将来的调整。
func loadOwnedAddress(repo addressReader, userId, addressId int64) (*model.UserAddress, error) {
	addr, err := repo.GetAddressById(addressId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.AddressNotExist
		}
		return nil, err
	}
	if addr.UserId != userId {
		return nil, model.UserNotSetAddress
	}
	return addr, nil
}

// addressReader 本文件用到的最小仓储能力。
//
// 声明成窄接口而不是直接用 *repository.AddressRepo:单测只需实现这几个
// 方法,不必构造 DB。这与 trade 侧 orderCanceller 的做法同一口径。
type addressReader interface {
	GetAddressById(addressId int64) (*model.UserAddress, error)
}

// bizErrors 地址域里**可预期**的业务失败。
//
// 白名单而不是黑名单:未知错误一律当基础设施故障上抛,让上游重试。
// 反过来(未知错误当业务失败)会把"user-service 抖了一下"变成
// "用户看到的地址不存在",而 DB 报错、连接断了都会走进那个分支。
var bizErrors = []error{
	model.AddressNotExist,
	model.AddressListIsNull,
	model.AddressNumsIsFull,
	model.UserNotSetAddress,
	model.ReceiverNotNull,
	model.PhoneMalformed,
}

// isBizError 判定是否是"重试也没用"的业务失败
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

// toProtoAddress 实体 → proto
func toProtoAddress(a *model.UserAddress) *v1_userv1.Address {
	if a == nil {
		return nil
	}
	return &v1_userv1.Address{
		AddressId:     a.AddressId,
		UserId:        a.UserId,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		PostalCode:    a.PostalCode,
		IsDefault:     a.IsDefault,
		AddressTag:    a.AddressTag,
		CreatedAt:     timestampOrNil(a.CreatedAt),
		UpdatedAt:     timestampOrNil(a.UpdatedAt),
	}
}

// toProtoAddressList 批量转换。
//
// **空列表返回空切片而不是 nil**:proto 的 repeated 字段两者都编码成
// "无元素",但空切片让调用方(None 语义)与测试断言更好写。
func toProtoAddressList(items []*model.UserAddress) []*v1_userv1.Address {
	out := make([]*v1_userv1.Address, 0, len(items))
	for _, a := range items {
		out = append(out, toProtoAddress(a))
	}
	return out
}

// toProtoSnapshot 实体 → 下单快照(只取发货需要的七个字段)
func toProtoSnapshot(a *model.UserAddress) *v1_userv1.AddressSnapshot {
	if a == nil {
		return nil
	}
	return &v1_userv1.AddressSnapshot{
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		PostalCode:    a.PostalCode,
	}
}

// timestampOrNil 零值时间转 nil。
//
// 不能直接 timestamppb.New(t):零值会变成公元 1 年,前端渲染出
// "0001-01-01",而语义上它是"没有这个时间"。
func timestampOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
