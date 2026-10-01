package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserProfileLogic {
	return &CreateUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateUserProfile 创建用户档案。同一用户已有档案时不重复创建。
func (l *CreateUserProfileLogic) CreateUserProfile(in *v1_userv1.CreateUserProfileReq) (*v1_userv1.CreateUserProfileResp, error) {
	userId := in.GetProfile().GetUserId()
	if userId == 0 {
		return &v1_userv1.CreateUserProfileResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	if _, err := l.svcCtx.UserRepo.GetUserById(userId); err != nil {
		if isNotFound(err) {
			return &v1_userv1.CreateUserProfileResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}

	existing, err := l.svcCtx.UserProfileRepo.GetProfileByUserId(userId)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return &v1_userv1.CreateUserProfileResp{Profile: converter.ToProtoUserProfile(existing)}, nil
	}

	profile := &model.UserProfile{
		UserId:    userId,
		Nickname:  in.GetProfile().GetNickname(),
		RealName:  in.GetProfile().GetRealName(),
		Gender:    in.GetProfile().GetGender(),
		AvatarURL: in.GetProfile().GetAvatarUrl(),
	}
	if err := l.svcCtx.UserProfileRepo.CreateProfile(profile); err != nil {
		return nil, err
	}
	return &v1_userv1.CreateUserProfileResp{Profile: converter.ToProtoUserProfile(profile)}, nil
}
