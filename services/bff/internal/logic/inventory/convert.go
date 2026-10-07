package inventory

import (
	"encoding/json"
	"errors"
	"time"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
// 每个 logic 包各有一份,理由见 role 包 helpers.go 的说明。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// ============================================================
// proto SkuStock → types.SkuInventory / types.SkuStockResp
// ============================================================
//
// 本域共三类差异(与 .api 顶部"product-service:类目/商品/库存"那段的
// ① ② ④ 对应):
//
//	改名  repeated items   → list(仅列表接口;单 SKU 是裸对象)
//	类型  spec_values      → JSON 文本变对象
//	推导  total_stock      → proto 没有这个字段,由 BFF 算
//
// 其余字段逐字段直传(单体 productclient.toModelSkuInventory 也是直传)。
//
// ------------------------------------------------------------
// total_stock:同一个名字,两种含义(本域最容易搞错的地方)
// ------------------------------------------------------------
//
// ① 单个 SKU(types.SkuInventory.TotalStock / types.SkuStockResp.TotalStock)
//
//	= **该 SKU 的** stock + lock_stock(可用 + 锁定)。
//
//	proto 的 SkuStock 只有 sku_id / sku_name / spu_id / spu_name /
//	spec_values / stock / lock_stock / sold_count / sku_status 九个字段,
//	**没有 total_stock**;而 HTTP 契约(SkuInventoryResp)里有这个字段。
//	故只能由 BFF 推导。
//
//	公式不是我挑的,是既有契约:单体 productclient.toModelSkuInventory
//	写的就是 `TotalStock: s.Stock + s.LockStock`,注释"与单体口径一致:
//	可售总量 = 可用 + 锁定"。BFF 照抄即为兼容。
//
// ② 某个 SPU 的聚合值(types.SpuStockResp.TotalStock)
//
//	= **该 SPU 下所有 SKU 的 stock 之和**(只累加可用,**不含锁定**),
//	由 product-service 算好给 BFF(inventoryservice.getskustocklistlogic
//	里是 totalStock += sku.Stock)。BFF 只透传,**绝不重算** ——
//	重算就得在 BFF 复制一份聚合口径,而两份口径必然漂移。
//
// 于是同一个响应里会出现"Σ 每一行的 total_stock ≠ 顶层 total_stock"
// (每行是可用+锁定,顶层只有可用)。这个不一致在单体时代就存在,
// BFF 不"顺手修正":前端库存页的"总库存/锁定/销量"是三个并列维度,
// 把顶层 total_stock 改成可用+锁定会把锁定数算两遍。
//
// 由此可得:顶层 total_stock + total_lock == Σ 每一行的 total_stock。

// totalStockOf 单个 SKU 的 total_stock 口径:可用 + 锁定。
//
// 抽成函数是为了让这个定义**只有一处** —— 单 SKU 响应与 SPU 列表
// 的每一行都要用它,散开写迟早有人只改一处,两边就此分叉。
func totalStockOf(p *v1_productv1.SkuStock) int64 {
	return p.GetStock() + p.GetLockStock()
}

// toSkuInventory proto 库存条目 → SPU 列表里的一行。
func toSkuInventory(p *v1_productv1.SkuStock) types.SkuInventory {
	if p == nil {
		return types.SkuInventory{}
	}
	return types.SkuInventory{
		SkuId:      p.GetSkuId(),
		SkuName:    p.GetSkuName(),
		SpuId:      p.GetSpuId(),
		SpuName:    p.GetSpuName(),
		SpecValues: parseSpecValues(p.GetSpecValues()),
		Stock:      p.GetStock(),
		LockStock:  p.GetLockStock(),
		TotalStock: totalStockOf(p),
		SoldCount:  p.GetSoldCount(),
		SkuStatus:  p.GetSkuStatus(),
	}
}

// toSkuStockResp proto 库存条目 → 单 SKU 的**裸对象**响应。
//
// 与 toSkuInventory 字段完全相同,但 goctl 不支持类型别名(.api 里
// 注释已说明:实测报 syntax error),types.SkuStockResp 是另一套独立
// 生成的结构体,故这里必须再映射一次 —— 不是复制粘贴的疏忽。
//
// p 为 nil 时返回 nil(而不是零值对象):下发的 nil 指针序列化成
// `data: null`,与单体 productclient.GetSkuStock → toModelSkuInventory
// 返回 nil 后 utils.Success(c, nil) 的形状一致。编造一个 sku_id=0
// 的对象更糟 —— 前端会把它当成一条真实库存渲染出来。
func toSkuStockResp(p *v1_productv1.SkuStock) *types.SkuStockResp {
	if p == nil {
		return nil
	}
	return &types.SkuStockResp{
		SkuId:      p.GetSkuId(),
		SkuName:    p.GetSkuName(),
		SpuId:      p.GetSpuId(),
		SpuName:    p.GetSpuName(),
		SpecValues: parseSpecValues(p.GetSpecValues()),
		Stock:      p.GetStock(),
		LockStock:  p.GetLockStock(),
		TotalStock: totalStockOf(p),
		SoldCount:  p.GetSoldCount(),
		SkuStatus:  p.GetSkuStatus(),
	}
}

// ============================================================
// proto StockLog → types.StockLogItem
// ============================================================
//
// **不映射 order_no 与 idempotency_key**,这不是漏了:
//
//	proto 的 StockLog 有这两个字段(16/17),而 types.StockLogItem
//	(取自单体 response.InventoryLogList = 内嵌 SysProductStockLog
//	+ sku_name/spu_name)里没有 —— 单体返回给前端的模型里就没有。
//
// 它们对下游是**追溯字段**(converter.ToProtoStockLog 的注释:新代码
// 用 order_no,order_id 只有历史流水有值),不是展示字段;而
// idempotency_key 更是内部幂等键,回给前端等于把它暴露成对外契约
// (前端一旦开始依赖,以后就删不掉了)。故照 HTTP 契约隐藏。
//
// created_at 走 formatTimestamp 格式化成 RFC3339 字符串(proto 那边是
// Timestamp,直接下发会序列化成 {"seconds":..,"nanos":..} 这种前端
// 认不出的形状)。
func toStockLogItem(p *v1_productv1.StockLog) types.StockLogItem {
	if p == nil {
		return types.StockLogItem{}
	}
	return types.StockLogItem{
		LogId:       p.GetLogId(),
		SkuId:       p.GetSkuId(),
		SkuName:     p.GetSkuName(),
		SpuName:     p.GetSpuName(),
		ChangeType:  p.GetChangeType(),
		ChangeQty:   p.GetChangeQty(),
		BeforeStock: p.GetBeforeStock(),
		AfterStock:  p.GetAfterStock(),
		BeforeLock:  p.GetBeforeLock(),
		AfterLock:   p.GetAfterLock(),
		OrderId:     p.GetOrderId(),
		Remark:      p.GetRemark(),
		CreatedAt:   formatTimestamp(p.GetCreatedAt()),
		CreateBy:    p.GetCreateBy(),
	}
}

// parseSpecValues 把 proto 的 spec_values(JSON 文本)解析成对象。
//
// 与 cart 域的 parseSpecValues 同一套做法(那份注释把理由写得更细,
// 这里只记库存域为什么也照做):
//
// 解析失败**不报错**,返回空对象 + 记一条 Error 日志 —— 这是一个
// 刻意的静默降级。spec_values 只是规格展示(如 {"颜色":"红"}),
// 它坏掉说明 product-service 那条数据有问题,但不该让**整个库存看板
// 打不开**:用户应当看到"这一行没显示规格",而不是白屏。
//
// 空对象({})而不是 nil:前端会做 spec_values.颜色 这类取值,nil 抛错
// 而 {} 安全,且 {} 序列化出来比 null 更接近"有规格但为空"。
//
// 代价是这类问题只出现在日志里 —— 故必须记。单体那边是
// productclient.textToJSONMap,行为相同(失败返回空 map)。
func parseSpecValues(s string) interface{} {
	// 空串先短路:proto 侧 spec_values 为空是**正常情况**(DB 该列
	// NOT NULL DEFAULT '{}',但历史行可能为空),走解析只会白白得到
	// 一条 "unexpected end of JSON input"。
	if s == "" {
		return map[string]interface{}{}
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		// **不打印原始值**:它是商品侧完全可控的输入,原样进日志
		// 会让日志格式被外部数据左右(超长串、换行)。只打长度。
		logx.Errorf("inventory: spec_values 不是合法 JSON(len=%d): %v", len(s), err)
		return map[string]interface{}{}
	}
	if m == nil {
		// `null` 是合法 JSON,Unmarshal 进 map 得到 nil,归一到 {}。
		return map[string]interface{}{}
	}
	return m
}

// ============================================================
// 时间:响应格式 与 查询参数解析
// ============================================================
//
// 与 order / payment 域同口径(order 的 helpers.go 有完整说明):
// 响应是 RFC3339 字符串(单体经 time.Time 序列化),查询参数容错
// 三种格式(gin 的 form 绑定比 encoding/json 宽松)。

// formatTimestamp timestamppb.Timestamp → RFC3339 字符串(nil 安全)。
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// parseTimeParam 把查询参数里的时间字符串解析成 Timestamp。
//
// 空串返回 nil(表示不筛选),不报错 —— 不传时间范围是正常的。
//
// **不在这里补"结束日的 23:59:59"**:proto 的注释写着"结束时间(含当日)",
// 那个语义由 product-service 实现(它的 repository 会把 end_time 截到当天
// 23:59:59)。BFF 自己加一天再加 23:59:59 会让两个服务各改一次,
// 同一天被算了两次。
//
// 格式非法返回错误(与 order / payment 一致)。注意这一类错误会落到
// 500 而不是 400:response.Failure 只把 response.ErrInvalidParam 认成
// 业务错。保持与既有域一致,不在这里单独搞一套。
func parseTimeParam(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t), nil
		}
	}
	return nil, errors.New("时间格式错误: " + s)
}

// optionalInt64 可空筛选条件:int64 → proto 的 Int64Value。
//
// .api 里 sku_id / spu_id 是 optional,不传即零值;proto 那边用包装类型
// 表达"不过滤"。故 0 必须翻译成 nil —— 传 &Int64Value{Value:0} 会被
// 服务端当成"筛 sku_id=0",查出空列表(比不筛更糟:页面显示"没有流水"
// 而不是"全部流水")。
//
// 只认 ==0 而不是 <=0:负数在单体那边会作为真实筛选值传下去(得到空
// 列表),把负数也归成"不过滤"是行为改变。id 恒正,实际不会遇到。
func optionalInt64(v int64) *wrapperspb.Int64Value {
	if v == 0 {
		return nil
	}
	return wrapperspb.Int64(v)
}
