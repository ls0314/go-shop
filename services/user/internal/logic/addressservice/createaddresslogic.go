package addressservicelogic

import (
	"context"
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// CreateAddressLogic 新增收货地址。
//
// 业务规则(照搬单体 address_service.go:34):
//   - 手机号格式 → 姓名/手机非空
//   - **首个地址强制设为默认**:否则会得到"有地址但没有默认地址"的状态,
//     而结算页要预选默认地址
//   - 指定默认时取消旧默认(保证默认唯一)
//   - 有效地址不超过 20 条
//
// 全部在一个事务内:默认地址的唯一性靠"先取消旧默认再插入新的"两步,
// 拆开会在两步之间留下"两个默认"或"零个默认"的窗口。
type CreateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAddressLogic {
	return &CreateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAddressLogic) CreateAddress(in *v1_userv1.CreateAddressReq) (*v1_userv1.CreateAddressResp, error) {
	if err := validateAddressFields(in.GetReceiverName(), in.GetReceiverPhone()); err != nil {
		return &v1_userv1.CreateAddressResp{ErrorMsg: err.Error()}, nil
	}

	addr := &model.UserAddress{
		UserId:        in.GetUserId(),
		ReceiverName:  in.GetReceiverName(),
		ReceiverPhone: in.GetReceiverPhone(),
		Province:      in.GetProvince(),
		City:          in.GetCity(),
		District:      in.GetDistrict(),
		DetailAddress: in.GetDetailAddress(),
		PostalCode:    in.GetPostalCode(),
		IsDefault:     in.GetIsDefault(),
		AddressTag:    in.GetAddressTag(),
	}

	var bizErr error
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repoTx := l.svcCtx.AddressRepo.WithTx(tx)

		// 上限校验。用 COUNT 而不是把地址簿拉出来数 ——
		// 这个动作不该依赖另一个查询的过滤条件与排序
		n, err := repoTx.CountAddress(addr.UserId)
		if err != nil {
			return err
		}
		if n >= model.AddressMaxCount {
			bizErr = model.AddressNumsIsFull
			return nil // 业务失败:回滚事务但不作为"基础设施错误"上抛
		}

		// 默认地址管理
		existing, err := repoTx.GetDefaultAddress(addr.UserId)
		switch {
		case err == nil:
			// 已有默认:仅在本次指定为默认时取消旧的
			if addr.IsDefault {
				if err := repoTx.UnsetDefault(existing.AddressId); err != nil {
					return err
				}
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 没有默认(含首个地址):强制设默认,忽略入参 ——
			// 让"总是有默认地址"成为不变量,而不是靠调用方记得传
			addr.IsDefault = true
		default:
			return err
		}

		return repoTx.CreateAddress(addr)
	})
	if err != nil {
		return nil, err
	}
	if bizErr != nil {
		return &v1_userv1.CreateAddressResp{ErrorMsg: bizErr.Error()}, nil
	}

	return &v1_userv1.CreateAddressResp{AddressId: addr.AddressId}, nil
}
