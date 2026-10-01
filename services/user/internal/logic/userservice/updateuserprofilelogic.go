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

type UpdateUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserProfileLogic {
	return &UpdateUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateUserProfile 局部更新档案。
// 入参的 ID 是用户ID(与档案主键不同):按 user_id 定位记录。
func (l *UpdateUserProfileLogic) UpdateUserProfile(in *v1_userv1.UpdateUserProfileReq) (*v1_userv1.UpdateUserProfileResp, error) {
	userId := in.GetUserProfileId()

	older, err := l.svcCtx.UserProfileRepo.GetProfileByUserId(userId)
	if err != nil {
		if isNotFound(err) {
			return &v1_userv1.UpdateUserProfileResp{ErrorMsg: model.UserProfileNotExist.Error()}, nil
		}
		return nil, err
	}

	newProfile := *older
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newProfile,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateUserProfileResp{ErrorMsg: err.Error()}, nil
	}

	// 归属关系不可通过更新改变
	newProfile.UserId = older.UserId
	newProfile.UserInfoID = older.UserInfoID

	if err := l.svcCtx.UserProfileRepo.UpdateProfile(&newProfile); err != nil {
		return nil, err
	}
	return &v1_userv1.UpdateUserProfileResp{Profile: converter.ToProtoUserProfile(&newProfile)}, nil
}
