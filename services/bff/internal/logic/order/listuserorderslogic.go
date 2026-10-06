// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	v1_tradev1 "demo-shop/api/gen/trade/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserOrdersLogic {
	return &ListUserOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListUserOrders 分页查询当前用户的订单。
//
// ============================================================
// 响应是 {list, total, page, page_size}
// ============================================================
//
// 单体 response.UserGetOrderListResp 就是这个形状 ——
// **注意 page_size 是 snake_case**,与用户列表的 pageSize(camelCase)
// 不同。这是既有的不一致(订单域用 request struct 的 form tag,
// 用户域用 c.DefaultQuery 手工拼),不要统一。
//
// ============================================================
// 列表项只有 8 个字段(proto 的 Order 有 16 个)
// ============================================================
//
// HTTP 列表项:order_id / order_no / order_status / total_amount /
//
//	pay_amount / detail_count / first_image / created_at
//
// proto Order 里**不在列表项**里的:user_id / pay_method / pay_time /
// address_snapshot / buyer_remark / expire_at / idempotent_key /
// updated_at。
//
// **列表项不含 address_snapshot** —— 那是详情页才要的。故这个列表
// 不需要解析 JSON,比详情轻。
//
// 同理没有 pay_method / pay_time:它们在"支付流程"里才有意义,
// 而订单列表只关心状态与金额。
//
// ============================================================
// 分页兜底并回写
// ============================================================
//
// .api 里 page / page_size 是 optional,不传时为 0,而 proto 传 0 会让
// 服务端行为不确定,故兜成 1/10(与单体的 DefaultQuery 一致)。
//
// 并把兜底后的值**回写到响应** —— 单体回的也是兜底后的值,
// 前端据此渲染分页控件(不传时看到 page=1、page_size=10)。
//
// ============================================================
// order_status 的取值由服务端定义
// ============================================================
//
// 空串表示不筛选(proto 注释)。前端订单页通常有"全部/待付款/
// 待发货/已完成"的 tab,切换时传对应 status。
//
// BFF **不校验也不转换** status —— 传错了会得到空列表(而不是报错),
// 那是服务端的行为。要在 BFF 拦就得维护一份状态枚举,而那份枚举
// 必然与服务端漂移。
func (l *ListUserOrdersLogic) ListUserOrders(req *types.ListUserOrdersReq) (*types.ListUserOrdersResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.OrderRPC.ListUserOrders(ctx, &v1_tradev1.ListUserOrdersReq{
		UserId:      userId,
		Page:        int32(page),
		PageSize:    int32(pageSize),
		OrderStatus: req.OrderStatus,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// make(..., 0, len) 而不是 var x []T:空时是 [] 而不是 null,
	// 前端 map/length 不会抛错。
	items := resp.GetItems()
	list := make([]types.UserOrderListItem, 0, len(items))
	for _, o := range items {
		list = append(list, types.UserOrderListItem{
			OrderId:     o.GetOrderId(),
			OrderNo:     o.GetOrderNo(),
			OrderStatus: o.GetOrderStatus(),
			TotalAmount: o.GetTotalAmount(),
			PayAmount:   o.GetPayAmount(),
			DetailCount: o.GetDetailCount(),
			FirstImage:  o.GetFirstImage(),
			CreatedAt:   formatTimestamp(o.GetCreatedAt()),
		})
	}

	return &types.ListUserOrdersResp{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
