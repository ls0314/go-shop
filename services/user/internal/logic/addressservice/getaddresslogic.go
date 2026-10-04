package addressservicelogic

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetAddressLogic 查询单个地址详情。
type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAddressLogic) GetAddress(in *v1_userv1.GetAddressReq) (*v1_userv1.GetAddressResp, error) {
	addr, err := loadOwnedAddress(l.svcCtx.AddressRepo, in.GetUserId(), in.GetAddressId())
	if err != nil {
		if isBizError(err) {
			return &v1_userv1.GetAddressResp{ErrorMsg: err.Error()}, nil
		}
		return nil, err
	}
	return &v1_userv1.GetAddressResp{Address: toProtoAddress(addr)}, nil
}
