package paymentservicelogic

import (
	"context"
	"time"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/trade/internal/converter"
	"demo-shop/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListPaymentsLogic 管理端支付流水列表的协议适配层
type ListPaymentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPaymentsLogic {
	return &ListPaymentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPaymentsLogic) ListPayments(in *v1_tradev1.ListPaymentsReq) (*v1_tradev1.ListPaymentsResp, error) {
	var startTime, endTime *time.Time
	if in.GetStartTime() != nil {
		t := in.GetStartTime().AsTime()
		startTime = &t
	}
	if in.GetEndTime() != nil {
		t := in.GetEndTime().AsTime()
		endTime = &t
	}

	items, total, page, pageSize, err := l.svcCtx.PaymentService.ListPayments(
		int(in.GetPage()), int(in.GetPageSize()),
		in.GetPayStatus(), in.GetPayMethod(), in.GetOrderNo(),
		startTime, endTime)
	if err != nil {
		return &v1_tradev1.ListPaymentsResp{ErrorMsg: bizErrMsg(err)}, err
	}
	return &v1_tradev1.ListPaymentsResp{
		Items: converter.ToProtoPaymentViews(items),
		Total: total,
		// 回显实际生效的分页参数(service 侧做过归一化与封顶)
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
