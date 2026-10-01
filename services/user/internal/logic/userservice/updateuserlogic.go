package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateUser 局部更新用户。
// 不处理口令:本接口不接收明文口令,也不重新哈希既有哈希值。
func (l *UpdateUserLogic) UpdateUser(in *v1_userv1.UpdateUserReq) (*v1_userv1.UpdateUserResp, error) {
	olderUser, err := l.svcCtx.UserRepo.GetUserById(in.UserId)
	if err != nil {
		if isNotFound(err) {
			return &v1_userv1.UpdateUserResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}

	newUser := *olderUser
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newUser,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateUserResp{ErrorMsg: err.Error()}, nil
	}

	// 不改口令:入参可能带 password 字段,统一还原为原哈希
	newUser.PasswordHash = olderUser.PasswordHash

	// 唯一性检查需排除自己,否则"保持原值"也会被判为冲突
	if newUser.Username != olderUser.Username {
		existing, err := l.svcCtx.UserRepo.GetUserByName(newUser.Username)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
		if existing != nil && existing.UserID != in.UserId {
			return &v1_userv1.UpdateUserResp{ErrorMsg: model.UsernameExist.Error()}, nil
		}
	}
	if newUser.Phone != olderUser.Phone && newUser.Phone != "" {
		existing, err := l.svcCtx.UserRepo.GetUserByPhone(newUser.Phone)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
		if existing != nil && existing.UserID != in.UserId {
			return &v1_userv1.UpdateUserResp{ErrorMsg: model.PhoneExist.Error()}, nil
		}
	}
	if newUser.Email != olderUser.Email && newUser.Email != "" {
		existing, err := l.svcCtx.UserRepo.GetUserByEmail(newUser.Email)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
		if existing != nil && existing.UserID != in.UserId {
			return &v1_userv1.UpdateUserResp{ErrorMsg: model.EmailExist.Error()}, nil
		}
	}

	if err := l.svcCtx.UserRepo.UpdateUser(&newUser); err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateUserResp{User: converter.ToProtoUser(&newUser)}, nil
}
