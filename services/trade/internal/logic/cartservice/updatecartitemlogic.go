package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCartItemLogic 改数量 / 改选中态的协议适配层
type UpdateCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartItemLogic {
	return &UpdateCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCartItemLogic) UpdateCartItem(in *v1_tradev1.UpdateCartItemReq) (*v1_tradev1.UpdateCartItemResp, error) {
	view, err := l.svcCtx.CartService.UpdateCartItem(l.ctx,
		in.GetCartItemId(), in.GetUserId(), in.GetQuantity(), in.GetIsSelected())
	if err != nil {
		return &v1_tradev1.UpdateCartItemResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.UpdateCartItemResp{Item: converter.ToProtoCartItem(view)}, nil
}
