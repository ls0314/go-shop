package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserProfileLogic {
	return &DeleteUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteUserProfile 按用户ID删除档案。无档案时按成功处理。
func (l *DeleteUserProfileLogic) DeleteUserProfile(in *v1_userv1.DeleteUserProfileReq) (*v1_userv1.DeleteUserProfileResp, error) {
	if err := l.svcCtx.UserProfileRepo.DeleteProfile(in.GetUserProfileId()); err != nil {
		return nil, err
	}
	return &v1_userv1.DeleteUserProfileResp{}, nil
}
