package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserProfileLogic) GetUserProfile(in *v1_userv1.GetUserProfileReq) (*v1_userv1.GetUserProfileResp, error) {
	profile, err := l.svcCtx.UserProfileRepo.GetProfileByUserId(in.UserId)
	if err != nil {
		if isNotFound(err) {
			return &v1_userv1.GetUserProfileResp{ErrorMsg: model.UserProfileNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_userv1.GetUserProfileResp{Profile: converter.ToProtoUserProfile(profile)}, nil
}
