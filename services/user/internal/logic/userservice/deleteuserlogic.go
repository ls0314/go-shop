package userservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *v1_userv1.DeleteUserReq) (*v1_userv1.DeleteUserResp, error) {
	if _, err := l.svcCtx.UserRepo.GetUserById(in.UserId); err != nil {
		if isNotFound(err) {
			return &v1_userv1.DeleteUserResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}

	// 有部门关联时不允许删除
	hasDeptRel, err := l.svcCtx.UserRepo.HasDeptRel(in.UserId)
	if err != nil {
		return nil, err
	}
	if hasDeptRel {
		return &v1_userv1.DeleteUserResp{ErrorMsg: model.UserHasDeptRel.Error()}, nil
	}

	// 有角色关联时不允许删除
	hasRoleRel, err := l.svcCtx.UserRepo.HasRoleRel(in.UserId)
	if err != nil {
		return nil, err
	}
	if hasRoleRel {
		return &v1_userv1.DeleteUserResp{ErrorMsg: model.UserHasRoleRel.Error()}, nil
	}

	if err := l.svcCtx.UserRepo.DeleteUser(in.UserId); err != nil {
		return nil, err
	}
	return &v1_userv1.DeleteUserResp{}, nil
}
