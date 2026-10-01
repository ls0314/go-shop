package userservicelogic

import (
	"context"
	"errors"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetSelfInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfInfoLogic {
	return &GetSelfInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSelfInfo 取当前登录用户的身份与档案。
// 档案缺失不算失败:注册时事务保证创建,但历史数据可能没有。
func (l *GetSelfInfoLogic) GetSelfInfo(in *v1_userv1.GetSelfInfoReq) (*v1_userv1.GetSelfInfoResp, error) {
	user, err := l.svcCtx.UserRepo.GetUserById(in.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_userv1.GetSelfInfoResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}

	var profile *model.UserProfile
	profile, err = l.svcCtx.UserProfileRepo.GetProfileByUserId(in.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &v1_userv1.GetSelfInfoResp{
		User:    converter.ToProtoUser(user),
		Profile: converter.ToProtoUserProfile(profile),
	}, nil
}
