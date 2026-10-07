// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package inventory

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListStockLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListStockLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListStockLogsLogic {
	return &ListStockLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListStockLogs 分页查询库存流水(管理端库存流水页)。
//
// ============================================================
// 请求与响应**都是 snake_case**(page_size,不是 pageSize)
// ============================================================
//
// 这是本域与 products / category 的分歧:那两个域的响应是 pageSize
// (camelCase,因为它们照抄单体的 gin.H{"pageSize": ...}),而库存流水
// 的响应取自单体 response.InventoryLogResp,是 page_size。
//
// 请求参数也是 snake_case —— 前端 InventoryLogQuery 的字段名。
// **不要"顺手统一"**:前端按 page_size 解,改名就是 undefined。
//
// ============================================================
// 分页兜底不在 BFF
// ============================================================
//
// 与 order / payment 列表不同(那两个域的单体 handler 用 DefaultQuery,
// 所以兜底搬到了 BFF),这里单体的 handler 只是把 page/page_size 原样
// 转发,默认值与封顶(1 / 50 / 上限 50)一直由服务端补。product-service
// 保留了这个行为,并把**生效后**的 page/page_size 回显在响应里。
//
// 故 BFF 既兜底也不夹紧,直接透传请求、回写服务端回显的值 —— 否则
// BFF 与服务端各有一份默认值/上限,改一处就会分叉。
//
// ============================================================
// 可空筛选条件
// ============================================================
//
// sku_id / spu_id 在 .api 里是 optional(零值即不过滤),proto 用
// Int64Value 包装,故经 optionalInt64 转成 nil(见 convert.go)。
//
// change_type 空串表示不筛(枚举值由服务端定义,BFF 不校验 —— 传错了
// 得到空列表,拦在 BFF 就得维护一份必然漂移的枚举)。
//
// start_time / end_time 三种格式容错解析;"含当日"的补时由服务端做。
func (l *ListStockLogsLogic) ListStockLogs(req *types.StockLogReq) (*types.StockLogResp, error) {
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

	resp, grpcErr := l.svcCtx.InventoryRPC.ListStockLogs(ctx, &v1_productv1.ListStockLogsReq{
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		SkuId:      optionalInt64(req.SkuId),
		SpuId:      optionalInt64(req.SpuId),
		ChangeType: req.ChangeType,
		StartTime:  startTime,
		EndTime:    endTime,
	})

	// ListStockLogsResp 有 error_msg(第 5 字段),排在 items/total/page/
	// page_size 之后 —— 抄其它域时容易漏掉它。
	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	items := resp.GetItems()
	list := make([]types.StockLogItem, 0, len(items))
	for _, it := range items {
		list = append(list, toStockLogItem(it))
	}

	return &types.StockLogResp{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     int(resp.GetPage()),
		PageSize: int(resp.GetPageSize()),
	}, nil
}
