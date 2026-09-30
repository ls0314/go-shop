package roleservice

import (
	"context"
	"demo-shop-back/src/model"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRoleLogic) GetRole(in *v1_userv1.GetRoleReq) (*v1_userv1.GetRoleResp, error) {
	role, err := l.svcCtx.RoleRepo.GetRoleById(in.RoleId)
	if err != nil {
		return &v1_userv1.GetRoleResp{ErrorMsg: model.RoleNotExist.Error()}, err
	}

	return &v1_userv1.GetRoleResp{Role: converter.ToProtoRole(role)}, nil
}
