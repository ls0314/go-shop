package addressservicelogic

import (
	"context"
	"errors"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// SetDefaultAddressLogic 设为默认地址。
//
// 在事务内"先取消旧默认、再设新默认"—— 两步必须原子。
// 拆开的话,并发两个请求同时设不同地址时都可能读到"旧默认是 A",
// 于是都去取消 A、再各自设自己,最后两个都是默认。
//
// 注意取消与设置都**不带 is_deleted 条件**:如果旧默认恰好是被软删
// 却漏了清标记的历史数据,这里要能把它清掉。
type SetDefaultAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDefaultAddressLogic) SetDefaultAddress(in *v1_userv1.SetDefaultAddressReq) (*v1_userv1.SetDefaultAddressResp, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repoTx := l.svcCtx.AddressRepo.WithTx(tx)

		addr, err := loadOwnedAddress(repoTx, in.GetUserId(), in.GetAddressId())
		if err != nil {
			return err
		}

		existing, err := repoTx.GetDefaultAddress(addr.UserId)
		switch {
		case err == nil:
			if existing.AddressId == addr.AddressId {
				// 已经是默认:幂等命中,不重复写。**不报错** ——
				// 用户连点两次"设为默认"不该看到失败
				return nil
			}
			if err := repoTx.UnsetDefault(existing.AddressId); err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 本来没有默认,直接设即可
		default:
			return err
		}

		return repoTx.SetDefault(addr.AddressId)
	})
	if err != nil {
		if isBizError(err) {
			return &v1_userv1.SetDefaultAddressResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}
	return &v1_userv1.SetDefaultAddressResp{}, nil
}
