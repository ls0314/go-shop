package addressservicelogic

import (
	"context"
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// DeleteAddressLogic 删除收货地址(软删除)。
//
// 若删掉的是默认地址,**自动把列表第一条顶上为默认** ——
// 列表按 `is_default DESC, updated_at DESC` 排,所以第一条就是
// "最近更新的那条有效地址"。这个兜底是必需的:没有它,用户删掉默认地址后
// 结算页就预选不出地址了,而他得自己再去点一次"设为默认"。
//
// 删掉最后一条地址时**不报错**:"没有默认地址"是合法状态
// (空地址簿),此时无事可做。
type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAddressLogic) DeleteAddress(in *v1_userv1.DeleteAddressReq) (*v1_userv1.DeleteAddressResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repoTx := l.svcCtx.AddressRepo.WithTx(tx)

		addr, err := loadOwnedAddress(repoTx, in.GetUserId(), in.GetAddressId())
		if err != nil {
			return err
		}

		if err := repoTx.DeleteAddress(addr.AddressId); err != nil {
			return err
		}

		if !addr.IsDefault {
			return nil
		}

		// 删的是默认地址:顶上最近更新的那条。
		// 注意软删已经生效,所以这次查询**看不到刚删的那条**
		list, err := repoTx.GetAddressList(addr.UserId)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			// 地址簿空了:没有默认地址是合法状态,不是错误。
			// 单体这里返回 AddressNotExist,会把"删掉最后一条地址"
			// 变成一个失败操作 —— 用户看到删除失败但地址其实已删
			return nil
		}
		return repoTx.SetDefault(list[0].AddressId)
	})
	if err != nil {
		if isBizError(err) {
			return &v1_userv1.DeleteAddressResp{ErrorMsg: err.Error()}, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_userv1.DeleteAddressResp{ErrorMsg: "地址不存在"}, nil
		}
		return nil, err
	}
	return &v1_userv1.DeleteAddressResp{}, nil
}
