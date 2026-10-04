package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListCartItemsLogic 的协议适配层
type ListCartItemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCartItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCartItemsLogic {
	return &ListCartItemsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCartItemsLogic) ListCartItems(in *v1_tradev1.ListCartItemsReq) (*v1_tradev1.ListCartItemsResp, error) {
	items, err := l.svcCtx.CartService.ListCartItems(in.GetUserId())
	if err != nil {
		return &v1_tradev1.ListCartItemsResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ListCartItemsResp{
		Items: converter.ToProtoCartItems(items),
	}, nil
}
