package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetCartItemCountLogic 购物车行数的协议适配层。
type GetCartItemCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCartItemCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartItemCountLogic {
	return &GetCartItemCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCartItemCountLogic) GetCartItemCount(in *v1_tradev1.GetCartItemCountReq) (*v1_tradev1.GetCartItemCountResp, error) {
	count, err := l.svcCtx.CartService.GetCartItemCount(in.GetUserId())
	if err != nil {
		return &v1_tradev1.GetCartItemCountResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetCartItemCountResp{Total: count}, nil
}
