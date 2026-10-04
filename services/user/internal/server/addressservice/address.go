package server

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	addressservicelogic "demo-shop/services/user/internal/logic/addressservice"
	"demo-shop/services/user/internal/svc"
)

// AddressServiceServer 收货地址服务的 gRPC 适配层(纯转发)。
//
// 与其它域的 server 同一约定:业务逻辑在 logic 层,这里只做协议转发。
// 错误约定也一致:
//   - 业务失败 → resp.ErrorMsg 有值,err == nil
//   - 基础设施失败 → err != nil
type AddressServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_userv1.UnimplementedAddressServiceServer
}

func NewAddressServiceServer(svcCtx *svc.ServiceContext) *AddressServiceServer {
	return &AddressServiceServer{svcCtx: svcCtx}
}

func (s *AddressServiceServer) CreateAddress(ctx context.Context, in *v1_userv1.CreateAddressReq) (*v1_userv1.CreateAddressResp, error) {
	return addressservicelogic.NewCreateAddressLogic(ctx, s.svcCtx).CreateAddress(in)
}

func (s *AddressServiceServer) GetAddress(ctx context.Context, in *v1_userv1.GetAddressReq) (*v1_userv1.GetAddressResp, error) {
	return addressservicelogic.NewGetAddressLogic(ctx, s.svcCtx).GetAddress(in)
}

func (s *AddressServiceServer) ListAddresses(ctx context.Context, in *v1_userv1.ListAddressesReq) (*v1_userv1.ListAddressesResp, error) {
	return addressservicelogic.NewListAddressesLogic(ctx, s.svcCtx).ListAddresses(in)
}

func (s *AddressServiceServer) UpdateAddress(ctx context.Context, in *v1_userv1.UpdateAddressReq) (*v1_userv1.UpdateAddressResp, error) {
	return addressservicelogic.NewUpdateAddressLogic(ctx, s.svcCtx).UpdateAddress(in)
}

func (s *AddressServiceServer) DeleteAddress(ctx context.Context, in *v1_userv1.DeleteAddressReq) (*v1_userv1.DeleteAddressResp, error) {
	return addressservicelogic.NewDeleteAddressLogic(ctx, s.svcCtx).DeleteAddress(in)
}

func (s *AddressServiceServer) SetDefaultAddress(ctx context.Context, in *v1_userv1.SetDefaultAddressReq) (*v1_userv1.SetDefaultAddressResp, error) {
	return addressservicelogic.NewSetDefaultAddressLogic(ctx, s.svcCtx).SetDefaultAddress(in)
}

func (s *AddressServiceServer) GetAddressSnapshot(ctx context.Context, in *v1_userv1.GetAddressSnapshotReq) (*v1_userv1.GetAddressSnapshotResp, error) {
	return addressservicelogic.NewGetAddressSnapshotLogic(ctx, s.svcCtx).GetAddressSnapshot(in)
}
