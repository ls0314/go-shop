package cartservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetCartPayPreviewLogic 的协议适配层
type GetCartPayPreviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCartPayPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartPayPreviewLogic {
	return &GetCartPayPreviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCartPayPreviewLogic) GetCartPayPreview(in *v1_tradev1.GetCartPayPreviewReq) (*v1_tradev1.GetCartPayPreviewResp, error) {
	items, totalQty, totalAmount, err := l.svcCtx.CartService.GetCartPayPreview(in.GetUserId())
	if err != nil {
		return &v1_tradev1.GetCartPayPreviewResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.GetCartPayPreviewResp{
		Preview: converter.ToProtoCartPayPreview(items, totalQty, totalAmount),
	}, nil
}
