// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package coupon

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCouponTemplatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCouponTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponTemplatesLogic {
	return &ListCouponTemplatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCouponTemplates 分页查询券模板(管理端)。
//
// ============================================================
// 响应是 {list, total, page, page_size}
// ============================================================
//
// proto 是 {items, total, page, page_size},HTTP 把 items 改名成 list,
// 其余键名保持 snake_case(见 convert.go 的列表说明)。
//
// **本域的 page_size 是 snake_case** —— 与 products / category 的
// pageSize(camelCase)相反。单体这里走的是 request struct 的 form tag
// (requset.GetCouponListReq),那两个域走的是 c.DefaultQuery("pageSize"),
// 是既有的不一致,不要统一。.api 顶部的"查询参数名是 camelCase(易错)"
// 那段就是为此写的:键名写错的后果是 httpx.Parse **静默丢弃**该参数,
// 筛选条件失效而列表照常返回,不报错。
//
// ============================================================
// 分页兜底并回写
// ============================================================
//
// .api 里 page / page_size 是 optional,不传时为 0,而 proto 传 0 会让
// 服务端行为不确定,故兜成 1/10、上限 100(与 order 域同一口径)。
//
// 并把兜底后的值**回写到响应**:前端据此渲染分页控件(不传时看到
// page=1、page_size=10)。proto 注释说服务端也会"回显实际生效的分页
// 参数",这里用的是**我们发出去的值** —— 两边的归一化口径相同
// (marketing-service 也是 1/10/100),故不打架。
//
// ============================================================
// 两个筛选参数都直传,不校验
// ============================================================
//
// coupon_name 是**精确匹配**(不是模糊),coupon_type 是枚举;
// 空串表示不过滤。传了不认识的值会得到空列表而不是报错 ——
// 那是服务端的行为,要在 BFF 拦就得维护一份类型枚举,而那份必然漂移。
//
// ============================================================
// status 是服务端推导的,直传
// ============================================================
//
// CouponTemplateItem.Status(not_started / ongoing / ended)由服务端按
// **它的**当前时间算,proto 里写明了理由:"由服务端算而非前端算 ——
// 两端各自判时间会因时钟漂移给出不同结论"。BFF 同理不重算:
// 重算的判据里还有 is_deleted(被软删的模板恒为 ended),而那个字段
// 根本不回给前端 —— 拿裁剪后的字段去推只会推出不同的结论。
func (l *ListCouponTemplatesLogic) ListCouponTemplates(req *types.AdminCouponListReq) (*types.AdminCouponListResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.ListCouponTemplates(ctx, &v1_marketingv1.ListCouponTemplatesReq{
		Page:       int32(page),
		PageSize:   int32(pageSize),
		CouponName: req.CouponName,
		CouponType: req.CouponType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.AdminCouponListResp{
		List:     toCouponTemplateItems(resp.GetItems()),
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
