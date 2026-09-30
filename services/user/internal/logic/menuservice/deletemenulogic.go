package menuservicelogic

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuLogic {
	return &DeleteMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteMenuLogic) DeleteMenu(in *v1_userv1.DeleteMenuReq) (out *v1_userv1.DeleteMenuResp, err error) {
	existing, _ := l.svcCtx.MenuRepo.GetMenuById(in.MenuId)
	if existing == nil {
		return &v1_userv1.DeleteMenuResp{ErrorMsg: model.MenuNotExist.Error()}, nil
	}

	// 有角色关联的菜单不允许删除
	hasRel, err := l.svcCtx.MenuRepo.CheckRoleRelMenu(in.MenuId)
	if err != nil {
		return nil, err
	}
	if hasRel {
		return &v1_userv1.DeleteMenuResp{ErrorMsg: model.MenuHasRel.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.MenuRepo.WithTx(tx).DeleteMenu(in.MenuId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.DeleteMenuResp{}, nil

}
