package menuservicelogic

import (
	"context"
	"demo-shop-back/src/model"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuLogic {
	return &GetMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuLogic) GetMenu(in *v1_userv1.GetMenuReq) (*v1_userv1.GetMenuResp, error) {
	menu, err := l.svcCtx.MenuRepo.GetMenuById(in.MenuId)
	if err != nil {
		return &v1_userv1.GetMenuResp{ErrorMsg: model.MenuNotExist.Error()}, nil
	}

	return &v1_userv1.GetMenuResp{Menu: converter.ToProtoMenu(menu)}, nil
}
