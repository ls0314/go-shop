package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetCartItemLogic 查单个 SKU 在购物车中的行的协议适配层
type GetCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartItemLogic {
	return &GetCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCartItemLogic) GetCartItem(in *v1_tradev1.GetCartItemReq) (*v1_tradev1.GetCartItemResp, error) {
	view, err := l.svcCtx.CartService.GetCartItem(l.ctx, in.GetUserId(), in.GetSkuId())
	if err != nil {
		return &v1_tradev1.GetCartItemResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetCartItemResp{Item: converter.ToProtoCartItem(view)}, nil
}
