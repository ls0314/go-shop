package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// SelectAllCartItemsLogic 全选 / 取消全选的协议适配层
type SelectAllCartItemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSelectAllCartItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SelectAllCartItemsLogic {
	return &SelectAllCartItemsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SelectAllCartItemsLogic) SelectAllCartItems(in *v1_tradev1.SelectAllCartItemsReq) (*v1_tradev1.SelectAllCartItemsResp, error) {
	affected, err := l.svcCtx.CartService.SelectAllCartItems(in.GetUserId(), in.GetIsSelected())
	if err != nil {
		return &v1_tradev1.SelectAllCartItemsResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.SelectAllCartItemsResp{Affected: affected}, nil
}
