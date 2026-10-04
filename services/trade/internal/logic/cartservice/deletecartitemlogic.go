package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteCartItemLogic 删除购物车行的协议适配层
type DeleteCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCartItemLogic {
	return &DeleteCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCartItemLogic) DeleteCartItem(in *v1_tradev1.DeleteCartItemReq) (*v1_tradev1.DeleteCartItemResp, error) {
	if err := l.svcCtx.CartService.DeleteCartItem(in.GetCartItemId(), in.GetUserId()); err != nil {
		return &v1_tradev1.DeleteCartItemResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.DeleteCartItemResp{}, nil
}
