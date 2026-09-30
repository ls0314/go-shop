package menuservicelogic

import (
	"context"
	"demo-shop-back/src/model"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateMenuLogic) UpdateMenu(in *v1_userv1.UpdateMenuReq) (*v1_userv1.UpdateMenuResp, error) {
	oldMenu, err := l.svcCtx.MenuRepo.GetMenuById(in.MenuId)
	if err != nil {
		return &v1_userv1.UpdateMenuResp{ErrorMsg: model.MenuNotExist.Error()}, nil
	}

	newMenu := *oldMenu
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newMenu,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateMenuResp{ErrorMsg: err.Error()}, nil
	}

	if newMenu.MenuName != oldMenu.MenuName {
		existing, _ := l.svcCtx.MenuRepo.GetMenuByUk(newMenu.ParentId, newMenu.MenuName)
		if existing != nil {
			return &v1_userv1.UpdateMenuResp{ErrorMsg: model.MenuExist.Error()}, nil
		}
	}

	if newMenu.ParentId != oldMenu.ParentId {
		sortId, err := l.svcCtx.MenuRepo.GetMaxSortId(newMenu.ParentId)
		if err != nil {
			return nil, err
		}
		newMenu.SortOrder = sortId
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.MenuRepo.WithTx(tx).UpdateMenu(&newMenu)
	})

	if err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateMenuResp{Menu: converter.ToProtoMenu(&newMenu)}, nil
}
