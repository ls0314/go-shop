package roleservice

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteRoleLogic) DeleteRole(in *v1_userv1.DeleteRoleReq) (*v1_userv1.DeleteRoleResp, error) {
	existing, _ := l.svcCtx.RoleRepo.GetRoleById(in.RoleId)
	if existing == nil {
		return &v1_userv1.DeleteRoleResp{ErrorMsg: model.RoleNotExist.Error()}, nil
	}

	if existing.IsSystem {
		return &v1_userv1.DeleteRoleResp{ErrorMsg: model.RoleIsSystem.Error()}, nil
	}

	userHasRel, err := l.svcCtx.RoleRepo.CheckRoleRelUser(in.RoleId)
	if err != nil {
		return nil, err
	}
	if userHasRel {
		return &v1_userv1.DeleteRoleResp{ErrorMsg: model.RoleHasUserRel.Error()}, nil
	}

	menuHasRel, err := l.svcCtx.RoleRepo.CheckRoleRelMenu(in.RoleId)
	if err != nil {
		return nil, err
	}
	if menuHasRel {
		return &v1_userv1.DeleteRoleResp{ErrorMsg: model.RoleHasMenuRel.Error()}, nil
	}

	permHasRel, err := l.svcCtx.RoleRepo.CheckRoleRelPerm(in.RoleId)
	if err != nil {
		return nil, err
	}
	if permHasRel {
		return &v1_userv1.DeleteRoleResp{ErrorMsg: model.RoleHasPermRel.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.RoleRepo.WithTx(tx).DeleteRole(in.RoleId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.DeleteRoleResp{}, nil

}
