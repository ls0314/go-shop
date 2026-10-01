package server

import (
	"context"

	"demo-shop/api/gen/user/v1"
	userservicelogic "demo-shop/services/user/internal/logic/userservice"
)

// ---- 认证 ----

func (s *UserServiceServer) Login(ctx context.Context, in *v1_userv1.LoginReq) (*v1_userv1.LoginResp, error) {
	l := userservicelogic.NewLoginLogic(ctx, s.svcCtx)
	return l.Login(in)
}
func (s *UserServiceServer) RefreshToken(ctx context.Context, in *v1_userv1.RefreshTokenReq) (*v1_userv1.RefreshTokenResp, error) {
	l := userservicelogic.NewRefreshTokenLogic(ctx, s.svcCtx)
	return l.RefreshToken(in)
}
func (s *UserServiceServer) Register(ctx context.Context, in *v1_userv1.RegisterReq) (*v1_userv1.RegisterResp, error) {
	l := userservicelogic.NewRegisterLogic(ctx, s.svcCtx)
	return l.Register(in)
}

// ---- 当前登录用户 ----

func (s *UserServiceServer) GetSelfInfo(ctx context.Context, in *v1_userv1.GetSelfInfoReq) (*v1_userv1.GetSelfInfoResp, error) {
	l := userservicelogic.NewGetSelfInfoLogic(ctx, s.svcCtx)
	return l.GetSelfInfo(in)
}
func (s *UserServiceServer) ListSelfPermCodes(ctx context.Context, in *v1_userv1.ListSelfPermCodesReq) (*v1_userv1.ListSelfPermCodesResp, error) {
	l := userservicelogic.NewListSelfPermCodesLogic(ctx, s.svcCtx)
	return l.ListSelfPermCodes(in)
}

// ---- 用户管理 ----

func (s *UserServiceServer) GetUser(ctx context.Context, in *v1_userv1.GetUserReq) (*v1_userv1.GetUserResp, error) {
	l := userservicelogic.NewGetUserLogic(ctx, s.svcCtx)
	return l.GetUser(in)
}
func (s *UserServiceServer) ListUsers(ctx context.Context, in *v1_userv1.ListUsersReq) (*v1_userv1.ListUsersResp, error) {
	l := userservicelogic.NewListUsersLogic(ctx, s.svcCtx)
	return l.ListUsers(in)
}
func (s *UserServiceServer) CreateUser(ctx context.Context, in *v1_userv1.CreateUserReq) (*v1_userv1.CreateUserResp, error) {
	l := userservicelogic.NewCreateUserLogic(ctx, s.svcCtx)
	return l.CreateUser(in)
}
func (s *UserServiceServer) UpdateUser(ctx context.Context, in *v1_userv1.UpdateUserReq) (*v1_userv1.UpdateUserResp, error) {
	l := userservicelogic.NewUpdateUserLogic(ctx, s.svcCtx)
	return l.UpdateUser(in)
}
func (s *UserServiceServer) DeleteUser(ctx context.Context, in *v1_userv1.DeleteUserReq) (*v1_userv1.DeleteUserResp, error) {
	l := userservicelogic.NewDeleteUserLogic(ctx, s.svcCtx)
	return l.DeleteUser(in)
}

// ---- 用户档案 ----

func (s *UserServiceServer) GetUserProfile(ctx context.Context, in *v1_userv1.GetUserProfileReq) (*v1_userv1.GetUserProfileResp, error) {
	l := userservicelogic.NewGetUserProfileLogic(ctx, s.svcCtx)
	return l.GetUserProfile(in)
}
func (s *UserServiceServer) CreateUserProfile(ctx context.Context, in *v1_userv1.CreateUserProfileReq) (*v1_userv1.CreateUserProfileResp, error) {
	l := userservicelogic.NewCreateUserProfileLogic(ctx, s.svcCtx)
	return l.CreateUserProfile(in)
}
func (s *UserServiceServer) UpdateUserProfile(ctx context.Context, in *v1_userv1.UpdateUserProfileReq) (*v1_userv1.UpdateUserProfileResp, error) {
	l := userservicelogic.NewUpdateUserProfileLogic(ctx, s.svcCtx)
	return l.UpdateUserProfile(in)
}
func (s *UserServiceServer) DeleteUserProfile(ctx context.Context, in *v1_userv1.DeleteUserProfileReq) (*v1_userv1.DeleteUserProfileResp, error) {
	l := userservicelogic.NewDeleteUserProfileLogic(ctx, s.svcCtx)
	return l.DeleteUserProfile(in)
}
