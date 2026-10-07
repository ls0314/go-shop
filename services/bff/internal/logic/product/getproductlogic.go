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

type GetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetProduct 管理端商品详情(含 SKU 列表与图片列表)。
//
// ============================================================
// 同一个 ProductDetail 要拆成两个 types 类型
// ============================================================
//
// proto 的管理端/用户端详情**共用一个 ProductDetail**(注释:"用户端只
// 消费 sku_list/image_list,其余字段一致"),而 HTTP 侧是两个类型:
//
//	AdminProductResp.SkuList []AdminSku   (含 cost_price / lock_stock)
//	UserProductResp.SkuList  []UserSku    (不含)
//
// 故 sku_list 的元素映射必须按端走(见 convert.go 的两份函数),
// 不能"共用一份再把多的字段抹掉"。
//
// ============================================================
// spec_template 是 JSON 文本 → 对象/数组
// ============================================================
//
// proto 的 spec_template 是 string,前端要的是可遍历的结构。
// 本项目的真实形状是**数组** [{name, values}],不能按对象解析 ——
// 否则一个合法模板会被判成解析失败并静默变成 {},前端编辑页的规格
// 编辑器变空,一保存就把模板整个丢了。见 parseSpecTemplate 的说明。
//
// ============================================================
// 查不到商品回 400,不是 404
// ============================================================
//
// 服务端把"商品不存在"放进 error_msg(业务失败),BFF 按既定分类回 400。
// 单体也是 500/400 一类的非 404(它的 handler 走 utils.Error)。
// **不要在这里改判 404** —— 那是契约变更,要单独做并通知前端。
func (l *GetProductLogic) GetProduct(req *types.ProductIdReq) (*types.AdminProductResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.GetProduct(ctx, &v1_productv1.GetProductReq{
		SpuId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 已软删 → 400
		return nil, err
	}

	return toAdminProductResp(resp.GetProduct()), nil
}
