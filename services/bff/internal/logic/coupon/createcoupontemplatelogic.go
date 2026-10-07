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

type CreateCouponTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCouponTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCouponTemplateLogic {
	return &CreateCouponTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateCouponTemplate 创建优惠券模板(管理端)。
//
// ============================================================
// 响应被**降级**成只有 template_id —— 与购物车的 CreateCartItem 同类
// ============================================================
//
// proto 的 CreateCouponTemplateResp 返回**完整实体**(注释写明
// "含服务端生成的 id 与 status",与 CreateScopeResp 同风格):
//
//	CreateCouponTemplateResp{template: {...}, error_msg}
//
// 而 HTTP 契约只有 template_id —— 单体 handler 把 proto 给的整行
// 压成了一个字段:
//
//	utils.Success(c, &response.CreateCouponResp{TemplateId: templateId})
//
// **BFF 保持这个降级**(前端零感知)。要升回整行是契约增强而非破坏性
// 变更(现有的 template_id 取值不变),但那是独立的一次改动,应当
// 单独做并通知前端 —— 不要夹在迁移里顺手改。
//
// 取值方式与单体一致(已核实 couponclient.CreateCouponTemplate 的最后
// 一行是 `return resp.Template.GetTemplateId(), "", nil`)。故若服务端
// 没填 template,两边都会得到 template_id=0 —— 不是 BFF 引入的差异,
// 但值得在端到端验证时确认一次:建完模板后响应里的 template_id 非 0。
//
// ============================================================
// 不需要 user_id
// ============================================================
//
// 这是管理端路由(挂 RequestMeta + Auth),券模板不属于任何个人,
// 故没有"从 JWT 取身份"的动作。同组的用户端三条才需要(见
// listusercouponslogic.go 的说明)。
//
// ============================================================
// 参数校验全在服务端
// ============================================================
//
// 券类型枚举、优惠力度 > 0、门槛 >= 0、总量 > 0、限领 <= 总量、
// "有效期二选一"(相对天数 or 固定起止)都在 marketing-service 的
// validateTemplate 里。BFF **不复制一份** —— 复制出来的那份必然与
// 服务端漂移,而且这里拦下只会让前端拿到一个措辞不同的错误。
// 唯一的例外是时间字符串的解析(见 toProtoCouponTemplate):
// 那是 HTTP → proto 的类型转换,服务端根本收不到字符串。
//
// 服务端拒绝时回 error_msg(11007 参数非法 / 11008 有效期配置非法),
// 经 Classify 落 ResultBiz → 400。
func (l *CreateCouponTemplateLogic) CreateCouponTemplate(req *types.CreateCouponTemplateReq) (*types.CreateCouponTemplateResp, error) {
	template, err := toProtoCouponTemplate(req)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CouponRPC.CreateCouponTemplate(ctx, &v1_marketingv1.CreateCouponTemplateReq{
		Template: template,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// 降级成那一个字段(见上方说明)。
	return &types.CreateCouponTemplateResp{
		TemplateId: resp.GetTemplate().GetTemplateId(),
	}, nil
}
