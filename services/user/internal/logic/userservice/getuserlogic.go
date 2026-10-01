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

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *v1_userv1.GetUserReq) (*v1_userv1.GetUserResp, error) {
	user, err := l.svcCtx.UserRepo.GetUserById(in.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &v1_userv1.GetUserResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}
	return &v1_userv1.GetUserResp{User: converter.ToProtoUser(user)}, nil
}
