package menuservicelogic

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateMenuLogic) CreateMenu(in *v1_userv1.CreateMenuReq) (*v1_userv1.CreateMenuResp, error) {
	menu := converter.FromProtoMenu(in.Menu)

	existing, _ := l.svcCtx.MenuRepo.GetMenuByUk(menu.ParentId, menu.MenuName)
	if existing != nil {
		return &v1_userv1.CreateMenuResp{ErrorMsg: model.MenuExist.Error()}, nil
	}

	if menu.SortOrder == 0 {
		sortId, err := l.svcCtx.MenuRepo.GetMaxSortId(menu.ParentId)
		if err != nil {
			return nil, err
		}
		menu.SortOrder = sortId
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.MenuRepo.WithTx(tx).CreateMenu(menu)
	})

	if err != nil {
		return nil, err
	}

	return &v1_userv1.CreateMenuResp{Menu: converter.ToProtoMenu(menu)}, nil
}
