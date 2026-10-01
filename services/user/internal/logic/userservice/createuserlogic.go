package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateUser 管理端创建用户。不建档案:档案由用户域自助接口维护。
func (l *CreateUserLogic) CreateUser(in *v1_userv1.CreateUserReq) (*v1_userv1.CreateUserResp, error) {
	if !ValidatePassword(in.Password) {
		return &v1_userv1.CreateUserResp{ErrorMsg: model.RegPasswordInvalid.Error()}, nil
	}

	username := in.GetUser().GetUsername()
	if username == "" {
		return &v1_userv1.CreateUserResp{ErrorMsg: model.UsernameExist.Error()}, nil
	}

	existing, err := l.svcCtx.UserRepo.GetUserByName(username)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return &v1_userv1.CreateUserResp{ErrorMsg: model.UsernameExist.Error()}, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	status := in.GetUser().GetStatus()
	if status == "" {
		status = "active"
	}

	user := &model.SysUser{
		Username:     username,
		PasswordHash: string(hash),
		Email:        in.GetUser().GetEmail(),
		Phone:        in.GetUser().GetPhone(),
		Status:       status,
	}
	if err := l.svcCtx.UserRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return &v1_userv1.CreateUserResp{User: converter.ToProtoUser(user)}, nil
}
