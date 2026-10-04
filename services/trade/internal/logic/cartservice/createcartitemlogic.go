package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateCartItemLogic 加购的协议适配层。
type CreateCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCartItemLogic {
	return &CreateCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCartItemLogic) CreateCartItem(in *v1_tradev1.CreateCartItemReq) (*v1_tradev1.CreateCartItemResp, error) {
	view, err := l.svcCtx.CartService.CreateCartItem(l.ctx, in.GetUserId(), in.GetSkuId(), in.GetQuantity())
	if err != nil {
		return &v1_tradev1.CreateCartItemResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.CreateCartItemResp{Item: converter.ToProtoCartItem(view)}, nil
}
