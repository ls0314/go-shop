package addressservicelogic

import (
	"context"
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateAddressLogic 局部更新收货地址。
//
// 用 FieldUpdate 列表而不是"可空字段":proto3 的标量无法区分"没传"与
// "传了零值",而地址的部分更新必须能区分 —— 否则一次只改手机号的请求
// 会把 is_default 一起抹成 false,于是用户莫名其妙失去了默认地址。
//
// 默认地址的唯一性在事务内维护:若本次把 is_default 从 false 改成 true,
// 先取消旧默认再保存。两步不能拆开,拆开会出现"两个默认"的窗口。
type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAddressLogic) UpdateAddress(in *v1_userv1.UpdateAddressReq) (*v1_userv1.UpdateAddressResp, error) {
	var updated *model.UserAddress

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repoTx := l.svcCtx.AddressRepo.WithTx(tx)

		old, err := loadOwnedAddress(repoTx, in.GetUserId(), in.GetAddressId())
		if err != nil {
			return err
		}

		// 用旧对象初始化,FieldUpdate 只覆盖传入的字段
		next := *old
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &next,
		})
		if err != nil {
			return err
		}
		if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
			return err
		}

		// 改了手机号/姓名就重新校验 —— 否则可以把手机号改成非法值,
		// 而创建时是校验过的(表上那条 CHECK 约束会以 500 的形式拒绝,
		// 那不是用户该看到的错误)
		if next.ReceiverPhone != old.ReceiverPhone || next.ReceiverName != old.ReceiverName {
			if err := validateAddressFields(next.ReceiverName, next.ReceiverPhone); err != nil {
				return err
			}
		}

		// 归属不可通过更新改写:它不是业务字段,是所有权
		next.UserId = old.UserId
		next.AddressId = old.AddressId

		// 默认地址唯一性:仅在"从非默认变为默认"时动旧默认
		if next.IsDefault && !old.IsDefault {
			existing, err := repoTx.GetDefaultAddress(old.UserId)
			switch {
			case err == nil:
				if err := repoTx.UnsetDefault(existing.AddressId); err != nil {
					return err
				}
			case errors.Is(err, gorm.ErrRecordNotFound):
				// 没有默认地址(理论上不该发生,因为创建时会强制设默认):
				// 本次直接设默认即可,不需要取消谁
			default:
				return err
			}
		}

		if err := repoTx.UpdateAddress(&next); err != nil {
			return err
		}
		updated = &next
		return nil
	})
	if err != nil {
		if isBizError(err) {
			return &v1_userv1.UpdateAddressResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}

	return &v1_userv1.UpdateAddressResp{Address: toProtoAddress(updated)}, nil
}
