package inventoryservicelogic

import (
	"context"
	"time"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/converter"
	"demo-shop/services/product/internal/repository"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// 流水分页上限,与单体 inventory_service.go 的 GetStockLogList 一致
const (
	stockLogDefaultPageSize = 50
	stockLogMaxPageSize     = 50
)

type ListStockLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListStockLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListStockLogsLogic {
	return &ListStockLogsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListStockLogs 分页查询库存流水(管理端库存流水页)。
//
// 从单体 src/service/inventory_service.go 的 GetStockLogList 平移。
// 分页默认值/封顶在服务端补,单体侧不再兜 —— 与库存看板三个查询同一口径。
func (l *ListStockLogsLogic) ListStockLogs(in *v1_productv1.ListStockLogsReq) (*v1_productv1.ListStockLogsResp, error) {
	page := int(in.Page)
	if page <= 0 {
		page = 1
	}
	// 防参数越界:<=0 用默认 50;>50 封顶 50(而非压成 50,避免大 pageSize 反而返回最少)
	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = stockLogDefaultPageSize
	}
	if pageSize > stockLogMaxPageSize {
		pageSize = stockLogMaxPageSize
	}

	q := repository.StockLogQuery{
		Page:       page,
		PageSize:   pageSize,
		SkuId:      optionalInt64(in.SkuId),
		SpuId:      optionalInt64(in.SpuId),
		ChangeType: in.ChangeType,
		StartTime:  optionalTime(in.StartTime),
		EndTime:    optionalTime(in.EndTime),
	}

	rows, total, err := l.svcCtx.InventoryLogRepo.ListStockLogs(q)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_productv1.StockLog, 0, len(rows))
	for i := range rows {
		items = append(items, converter.ToProtoStockLog(&rows[i]))
	}

	return &v1_productv1.ListStockLogsResp{
		Items:    items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// optionalInt64 把 proto 的可选 int64 转成查询条件;未传返回 nil 表示不过滤。
// 只服务本 RPC,故随本文件而不进 helper。
func optionalInt64(v *wrapperspb.Int64Value) *int64 {
	if v == nil {
		return nil
	}
	val := v.GetValue()
	return &val
}

// optionalTime 把 proto 的 Timestamp 转成查询条件;未传返回 nil。
// 返回指针而非值:零值时间会被误当成"1970 年"这个真实边界。
func optionalTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
