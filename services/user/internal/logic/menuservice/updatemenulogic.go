package menuservicelogic

import (
	"context"
	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"
	"errors"

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

	// 名字或父节点任一变化都可能在新父节点下撞名,故两者都变才跳过判重
	if newMenu.MenuName != oldMenu.MenuName || newMenu.ParentId != oldMenu.ParentId {
		existing, err := l.svcCtx.MenuRepo.GetMenuByUk(newMenu.ParentId, newMenu.MenuName)
		// 只换父节点不改名时,查回的是自身,不算冲突
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.MenuId != in.MenuId {
			return &v1_userv1.UpdateMenuResp{ErrorMsg: model.MenuExist.Error()}, nil
		}
	}

	if newMenu.ParentId != oldMenu.ParentId {
		sortId, err := l.svcCtx.MenuRepo.GetMaxSortId(newMenu.ParentId)
		if err != nil {
			return nil, err
		}
		// 挪到新父节点末尾
		newMenu.SortOrder = sortId + 1
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.MenuRepo.WithTx(tx).UpdateMenu(&newMenu)
	})

	if err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateMenuResp{Menu: converter.ToProtoMenu(&newMenu)}, nil
}
