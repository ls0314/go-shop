package addressservicelogic

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListAddressesLogic 查询用户的收货地址列表。
//
// **空列表是正常状态**(新用户还没填地址),返回空数组而不是错误。
// 单体的 GetAddressList 把"没有地址"返回成 AddressListIsNull,由调用方
// 忽略它 —— 那是把正常状态包装成错误再让所有人都记得忽略,
// 迁出时改掉了。注意 model.AddressListIsNull 仍保留:它是跨服务文案契约的
// 一部分,只是不再被本方法产出。
//
// 排序由仓储保证:默认地址最前,其余按 updated_at 倒序。
type ListAddressesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAddressesLogic) ListAddresses(in *v1_userv1.ListAddressesReq) (*v1_userv1.ListAddressesResp, error) {
	list, err := l.svcCtx.AddressRepo.GetAddressList(in.GetUserId())
	if err != nil {
		return nil, err
	}
	return &v1_userv1.ListAddressesResp{Items: toProtoAddressList(list)}, nil
}
