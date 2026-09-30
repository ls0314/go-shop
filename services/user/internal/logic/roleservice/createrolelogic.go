package roleservice

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRoleLogic) CreateRole(in *v1_userv1.CreateRoleReq) (*v1_userv1.CreateRoleResp, error) {
	role := converter.FromProtoRole(in.Role)

	existing, _ := l.svcCtx.RoleRepo.GetRoleById(role.RoleId)
	if existing != nil {
		return &v1_userv1.CreateRoleResp{ErrorMsg: model.RoleExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.RoleRepo.WithTx(tx).CreateRole(role)
	})

	if err != nil {
		return nil, err
	}

	return &v1_userv1.CreateRoleResp{Role: converter.ToProtoRole(role)}, nil
}
