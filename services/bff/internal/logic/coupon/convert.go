package coupon

import (
	"errors"
	"time"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/bff/internal/types"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// errNoIdentity 取不到调用方身份(路由未挂 Auth 中间件 = 配置错误)。
// 每个 logic 包各有一份,理由见 role 包 helpers.go 的说明。
var errNoIdentity = errors.New("取不到调用方身份:该路由未挂 Auth 中间件")

// ============================================================
// proto → types 的转换。本域共三类差异(与 .api 里优惠券那段的注释对应):
//
//	改名  repeated items → list(列表接口)
//	      proto 的 expire_at 在领券响应里叫 expire_time(见
//	      ReceiveCoupon 的说明),在"我的券"里仍叫 expire_at
//	拍平  CouponTemplateForUser{template, held_count, remaining_count}
//	      → types.UserCouponTemplateItem 的一行平字段
//	裁剪  is_deleted / created_at / updated_at / template_id / user_id
//	      不回给前端(单体在这些字段上标了 json:"-" 或模型里就没有)
//
// 另外两类是**刻意不做**的:status 与 pay_after 虽然 proto 与 HTTP
// 两边都有,但都是服务端算出来的,这里原样搬(理由见各 logic)。
//
// 时间字段一律走 formatTimestamp 变成字符串 —— proto 的 Timestamp
// 直接下发会序列化成 {"seconds":..,"nanos":..},前端认不出。
// ============================================================

// ============================================================
// 列表:proto 的 items → HTTP 的 list
// ============================================================
//
// 单体 handler 手工把 resp.Items 改名成 list(12 个列表接口全部如此,
// 见 bff.api 顶部①),故 types.*Resp 里都是 `List []X `json:"list"``。
// 照 proto 字段名写 items,前端拿到的是 undefined —— **而且不报错**。
//
// 空列表必须是 **[]** 而不是 null(故用 make(..., 0, len)),前端有
// list.length / list.map() 这类写法,nil 会直接抛错。
//
// nil 元素**跳过**,与单体 couponclient.convert.go 的每个循环一致
// (`if it == nil { continue }`):一行 user_coupon_id=0 / coupon_name=""
// 的"券"会被前端当成一张真券渲染出来,而少一行至少不会被误认。
// 领券中心那类嵌套结构还额外要求 template 非 nil —— 它这一行的每个
// 展示字段都取自嵌套的 template,模板缺失时整行没有意义。

// toCouponTemplateItems proto 模板 → 管理端模板列表。
func toCouponTemplateItems(ps []*v1_marketingv1.CouponTemplate) []types.CouponTemplateItem {
	out := make([]types.CouponTemplateItem, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			continue
		}
		out = append(out, types.CouponTemplateItem{
			TemplateId:      p.GetTemplateId(),
			CouponName:      p.GetCouponName(),
			CouponType:      p.GetCouponType(),
			ThresholdAmount: p.GetThresholdAmount(),
			DiscountAmount:  p.GetDiscountAmount(),
			TotalCount:      p.GetTotalCount(),
			ReceivedCount:   p.GetReceivedCount(),
			PerUserLimit:    p.GetPerUserLimit(),
			UsableDays:      p.GetUsableDays(),
			StartTime:       formatTimestamp(p.GetStartTime()),
			EndTime:         formatTimestamp(p.GetEndTime()),
			// 直传不重算:not_started / ongoing / ended 由服务端按
			// **它的**当前时间推导(proto 注释:"两端各自判时间会因
			// 时钟漂移给出不同结论"),理由见 ListCouponTemplates。
			Status: p.GetStatus(),
		})
	}
	return out
}

// toUserCouponItems proto 券实例 → "我的卡券"列表。
//
// 不映射的字段:template_id(HTTP 模型里没有)、user_id(单体 json:"-")、
// created_at(排序用,不展示)。**expire_at 与 used_at 都要格式化** ——
// used_at 对未使用的券是 nil,格式化成空串(单体那边是 time.Time 的零值,
// 见 formatTimestamp 的说明)。
func toUserCouponItems(ps []*v1_marketingv1.UserCoupon) []types.UserCouponItem {
	out := make([]types.UserCouponItem, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			continue
		}
		out = append(out, types.UserCouponItem{
			UserCouponId:    p.GetUserCouponId(),
			CouponName:      p.GetCouponName(),
			CouponType:      p.GetCouponType(),
			ThresholdAmount: p.GetThresholdAmount(),
			DiscountAmount:  p.GetDiscountAmount(),
			// status 是库里 user_coupon.status(unused/used/expired),
			// 与模板的 status 是两套枚举 —— 不要互相套用。
			Status:   p.GetStatus(),
			ExpireAt: formatTimestamp(p.GetExpireAt()),
			OrderNo:  p.GetOrderNo(),
			UsedAt:   formatTimestamp(p.GetUsedAt()),
		})
	}
	return out
}

// toUserCouponTemplateItems proto 领券中心项 → 用户端模板列表。
//
// **把嵌套的 template 拍平**:proto 的 CouponTemplateForUser 是
// {template, held_count, remaining_count},而 HTTP 的行是扁平的
// (types.UserCouponTemplateItem)。单体 couponclient 也是这么拍平的
// (它逐字段拷 tpl.*),故前端读的是 row.template_id 而不是
// row.template.template_id。
//
// 不映射 template.status:HTTP 契约里没有这个字段。领券中心的模板由
// 服务端按有效期过滤过(usable_days > 0 OR end_time > NOW()),列表里
// 不存在"已结束"的行,前端也没有这个角标要渲染。
func toUserCouponTemplateItems(
	ps []*v1_marketingv1.CouponTemplateForUser,
) []types.UserCouponTemplateItem {
	out := make([]types.UserCouponTemplateItem, 0, len(ps))
	for _, p := range ps {
		if p == nil || p.GetTemplate() == nil {
			continue
		}
		tpl := p.GetTemplate()
		out = append(out, types.UserCouponTemplateItem{
			TemplateId:      tpl.GetTemplateId(),
			CouponName:      tpl.GetCouponName(),
			CouponType:      tpl.GetCouponType(),
			ThresholdAmount: tpl.GetThresholdAmount(),
			DiscountAmount:  tpl.GetDiscountAmount(),
			PerUserLimit:    tpl.GetPerUserLimit(),
			// 这两个是服务端按 user_id 算的:已领数与剩余量。
			// 前端据此在 held_count >= per_user_limit 时置灰"领取"按钮,
			// 故 BFF 必须原样透传,不能自己数。
			HeldCount:      p.GetHeldCount(),
			RemainingCount: p.GetRemainingCount(),
			UsableDays:     tpl.GetUsableDays(),
			StartTime:      formatTimestamp(tpl.GetStartTime()),
			EndTime:        formatTimestamp(tpl.GetEndTime()),
		})
	}
	return out
}

// toAvailableCouponItems proto 可用券 → 结算页列表。
//
// 字段逐个直传,**含 pay_after**:那是服务端算好的"用这张券之后的
// 应付金额"(见 ListAvailableCoupons 的说明)。
func toAvailableCouponItems(ps []*v1_marketingv1.AvailableCoupon) []types.AvailableCouponItem {
	out := make([]types.AvailableCouponItem, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			continue
		}
		out = append(out, types.AvailableCouponItem{
			UserCouponId:    p.GetUserCouponId(),
			CouponName:      p.GetCouponName(),
			CouponType:      p.GetCouponType(),
			ThresholdAmount: p.GetThresholdAmount(),
			DiscountAmount:  p.GetDiscountAmount(),
			PayAfter:        p.GetPayAfter(),
		})
	}
	return out
}

// ============================================================
// types → proto:创建模板的请求体
// ============================================================

// toProtoCouponTemplate HTTP 请求体(扁平字段)→ proto 的 CouponTemplate。
//
// 单体那边是 gin 把 JSON 绑到 requset.CreateCouponReq,再由
// couponclient.CreateCouponTemplate 组装成 proto 实体 —— 组装这一步
// 在 BFF 就是这里。
//
// **start_time / end_time 是字符串而 proto 侧是 Timestamp**,必须解析:
//
//	空串 → nil,表示"未配置固定有效期"(相对有效期模式)
//	非空 → parseTimeParam(与 order / inventory 域同口径)
//
// nil 而不是零值 Timestamp 是关键:marketing-service 的
// FromProtoCouponTemplate 按 `p.StartTime != nil` 决定要不要赋值,
// 故 nil 才能表达"没配"。若传一个零值 Timestamp,0001-01-01 会被当成
// 一个合法的过去时间,validateTemplate 的"有效期二选一"就走到另一条
// 分支去了 —— 单体的 `if !req.StartTime.IsZero()` 正是为了避开这个。
//
// 前端(admin CouponForm.vue)相对有效期模式下**根本不发**这两个字段,
// 固定有效期模式下发的是 el-date-picker 的
// value-format="YYYY-MM-DDTHH:mm:ssZ",即 RFC3339。
//
// 其余字段逐个直传。received_count 不在入参里:它是运营不能自填的
// 计数器,由服务端起于 0、领取时原子自增。
func toProtoCouponTemplate(req *types.CreateCouponTemplateReq) (*v1_marketingv1.CouponTemplate, error) {
	startTime, err := parseTimeParam(req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := parseTimeParam(req.EndTime)
	if err != nil {
		return nil, err
	}

	return &v1_marketingv1.CouponTemplate{
		CouponName:      req.CouponName,
		CouponType:      req.CouponType,
		ThresholdAmount: req.ThresholdAmount,
		DiscountAmount:  req.DiscountAmount,
		TotalCount:      req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		UsableDays:      req.UsableDays,
		StartTime:       startTime,
		EndTime:         endTime,
	}, nil
}

// ============================================================
// 时间:响应格式 与 请求参数解析
// ============================================================
//
// 与 order / inventory 域同口径(order 的 helpers.go 有完整说明):
// proto 是 Timestamp,HTTP 侧是字符串。

// formatTimestamp timestamppb.Timestamp → RFC3339 字符串(nil 安全)。
//
// 为什么是 UTC:单体的模型字段是 time.Time,gin 用本地时区序列化
// (DS-A-13 的样例数据里是 "2026-08-05T14:30:00+08:00"),BFF 统一成
// UTC —— 表示的是**同一时刻**,前端 new Date(...) 两种都认。
//
// nil → 空串。**但空串并不代表"没有值"**,本域有个反例:
// marketing 的 ToProtoCouponTemplate 是无条件 timestamppb.New(t.StartTime),
// 故相对有效期(usable_days>0)的模板带的是**零值 Timestamp 而非 nil**,
// 这里会格式化成 "0001-01-01T00:00:00Z"。这正是为什么前端在
// usable_days>0 时只渲染"领取后 N 天有效",不去看 start_time / end_time。
//
// 与单体的一个可见差异:单体那一层是 time.Time(gin 直接序列化),
// 未使用的券 used_at 为零值,出来是 "0001-01-01T00:00:00Z"
// (DS-A-13 的样例数据就是这么写的);BFF 这里给空串。前端只渲染
// expire_at、没读 used_at,故这个差异没有实际影响 —— 但换实现的人
// 应当知道它存在,别把它当成"BFF 漏填了字段"。
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// parseTimeParam 把请求里的时间字符串解析成 Timestamp。
//
// 空串返回 nil(表示未配置),不报错 —— 相对有效期模式下不传这两个
// 字段是正常的。
//
// 容错三种格式(与 order / inventory 一致):单体在"创建模板"这条上走
// JSON 绑定(time.Time),只认 RFC3339;而 order / inventory 的查询参数
// 走 gin 的 form 绑定,容错更宽(还有 "2006-01-02 15:04:05" 与
// "2006-01-02")。同一个 BFF 里两套口径并存更糟,故统一取宽的那套 ——
// 前端的 el-date-picker 给的就是 RFC3339,宽出来的两种只是容错。
//
// 格式非法返回错误。注意这类错误会落到 500 而不是 400:
// response.Failure 只把 response.ErrInvalidParam 认成业务错。
// 保持与既有域一致,不在这里单独搞一套。
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

// ============================================================
// 分页兜底
// ============================================================
//
// 与 order 域同口径(那里的注释有完整推导):
//
//	不传(page<=0)      → 1 / 10
//	page_size > 100     → 夹到 100
//
// **必须兜底**:.api 的分页字段是 optional int,不传时为 0,而 proto 传 0
// 会让服务端行为不确定。本域的参数名是 page_size(snake_case,与 order
// 域相同),与 products / category 的 pageSize(camelCase)相反 —— 别混。
//
// marketing-service 的同名兜底也是 1 / 10 / 100,故这里兜过的值就是
// 服务端实际生效的值,回写到响应不会与它回显的值打架。
const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePage 兜底并夹紧分页参数。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
