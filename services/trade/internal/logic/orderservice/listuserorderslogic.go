package orderservicelogic

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListUserOrdersLogic 用户端订单列表的协议适配层
type ListUserOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserOrdersLogic {
	return &ListUserOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserOrdersLogic) ListUserOrders(in *v1_tradev1.ListUserOrdersReq) (*v1_tradev1.ListUserOrdersResp, error) {
	page, err := l.svcCtx.OrderService.ListUserOrders(in.GetUserId(), converter.FromProtoListUserOrders(in))
	if err != nil {
		return &v1_tradev1.ListUserOrdersResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ListUserOrdersResp{
		Items: converter.ToProtoOrderList(page.Items),
		Total: page.Total,
		// 回显**实际生效**的分页参数(service 侧做过归一化与封顶),
		// 前端据此校准页码控件
		Page:     int32(page.Page),
		PageSize: int32(page.PageSize),
	}, nil
}
