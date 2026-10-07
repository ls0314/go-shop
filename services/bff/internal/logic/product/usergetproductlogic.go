// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserGetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserGetProductLogic {
	return &UserGetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserGetProduct 用户端商品详情(仅已上架且未删除)。
//
// ============================================================
// sku_list 必须走 UserSku,不能复用管理端那份
// ============================================================
//
// proto 两端共用一个 ProductDetail,而 HTTP 侧是两个类型:
//
//	AdminProductResp.SkuList []AdminSku   含 cost_price / lock_stock
//	UserProductResp.SkuList  []UserSku    不含
//
// 对 C 端返回 cost_price 等于把**毛利**公开 —— 抓一次详情接口就能算出
// 每个 SKU 的成本。同理 lock_stock(下单未支付的锁定量)属于内部库存
// 口径,不该出现在前台。故这里走 toUserSkus,它按 UserSku 逐字段取,
// **不是**"取了再删"。
//
// ============================================================
// 详情对下架商品返回 400(而不是 404 / 空对象)
// ============================================================
//
// 服务端 loadUserDetail 要求 spu_status = published 且未删除,否则回
// error_msg"商品不存在" → BFF 回 400。
//
// 注意**文案与真实原因不同**:下架商品、已删商品、本来就不存在的 id
// 拿到的是同一句"商品不存在"。这是刻意的(不向上游暴露"这个商品存在
// 但下架了"),故前端不要试图从文案里区分这三种情况。
//
// 场景提醒:用户把商品页开着,运营期间下架 → 用户点"加入购物车"前的
// 详情刷新会拿到 400;此时前端该提示"商品已下架",而不是"网络错误"。
//
// ============================================================
// spec_template 也是 JSON 文本 → 数组/对象
// ============================================================
//
// 用户端也要规格模板(选购时要按"颜色/尺码"渲染可选项),故同样走
// parseSpecTemplate(它不做形状假设,数组才不会被解析坏)。
func (l *UserGetProductLogic) UserGetProduct(req *types.ProductIdReq) (*types.UserProductResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.UserGetProduct(ctx, &v1_productv1.UserGetProductReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 已软删 / 未上架 → 400(文案统一为"商品不存在")
		return nil, err
	}

	return toUserProductResp(resp.GetProduct()), nil
}
