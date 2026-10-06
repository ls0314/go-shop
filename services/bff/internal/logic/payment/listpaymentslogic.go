// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPaymentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPaymentsLogic {
	return &ListPaymentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListPayments 分页查询支付记录(管理端,不判权)。
//
// 响应元素对应单体 response.GetPaymentList,含 username / order_no
// (服务端联表查出来的)。
//
// 分页默认值是 **20/最大100**(与订单列表的 10 不同),见 helpers.go。
//
// 时间筛选容错三种格式(gin 的 form 绑定比 encoding/json 宽松)。
func (l *ListPaymentsLogic) ListPayments(req *types.ListPaymentsReq) (*types.ListPaymentsResp, error) {
	page, pageSize := normalizePaymentPage(req.Page, req.PageSize)

	startTime, err := parseTimeParam(req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := parseTimeParam(req.EndTime)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.PaymentRPC.ListPayments(ctx, &v1_tradev1.ListPaymentsReq{
		Page:      int32(page),
		PageSize:  int32(pageSize),
		PayStatus: req.PayStatus,
		PayMethod: req.PayMethod,
		OrderNo:   req.OrderNo,
		StartTime: startTime,
		EndTime:   endTime,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	items := resp.GetItems()
	list := make([]types.AdminPaymentListItem, 0, len(items))
	for _, p := range items {
		list = append(list, types.AdminPaymentListItem{
			PaymentId: p.GetPaymentId(),
			PayNo:     p.GetPayNo(),
			OrderNo:   p.GetOrderNo(),
			// 用户名快照(建支付时从订单读)—— 不需要查 user 域
			Username:  p.GetUsername(),
			PayMethod: p.GetPayMethod(),
			PayAmount: p.GetPayAmount(),
			PayStatus: p.GetPayStatus(),
			PayTime:   formatTimestamp(p.GetPayTime()),
		})
	}

	return &types.ListPaymentsResp{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
