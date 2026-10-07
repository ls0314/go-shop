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

type UpdateProductFullLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductFullLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFullLogic {
	return &UpdateProductFullLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateProductFull 整体更新:**以请求为准**覆盖 SPU 基本字段 + SKU 列表 + 图片列表。
//
// ============================================================
// SKU 对比规则(单体 requset.FullUpdateProductReq 的注释原文)
// ============================================================
//
//	传入列表中有 sku_id 的 → 更新
//	无 sku_id 的          → 新增
//	DB 有但列表里没有的    → **删除**(软删)
//
// 第三条就是"整体更新会删数据"的原因,也是这条接口与 UpdateProduct
// (局部更新)最容易被混用的地方:
//
//	前端只想改个价格,却把带全量语义的这条接口当"保存按钮"用;
//	而它提交的 sku_list 若来自一个**不完整**的数据源(比如只渲染了
//	第一页、或某个 SKU 被前面的校验过滤掉了),那些没出现在请求里的
//	SKU 就会被静默软删 —— 商品页上凭空少几个规格,且没有任何报错。
//
// 规则**由服务端执行**(它按 keepIds 软删差集),BFF 只负责把请求里
// 的 sku_list 原样映射下去,**不改也不补** —— 任何"帮你留一手"的
// 兜底(比如把请求里没有的 sku_id 也塞回去)都会让这份语义变得不可
// 预测,而且 BFF 根本不知道 DB 里有哪些 SKU(那要跨服务查)。
//
// ============================================================
// proto 的 UpdateProductFullReq 是**扁平**的,不是 {spu_id, spu}
// ============================================================
//
// 与 CreateProductReq(嵌套一个 Spu)不同,这条 RPC 把字段平铺在
// 请求体上,并用 wrappers 区分"没传"与"传了 0":
//
//	category_id / priority  → *Int64Value(服务端 nil 即不改)
//	其余 string             → 空串即不改(服务端逐个 if != "" 判断)
//	sku_list / image_list   → len > 0 才处理
//	delete_image_ids        → 图片的**显式删除**通道
//
// 故这里的映射是"逐字段搬运",没有组装步骤。
//
// ============================================================
// delete_image_ids:原来少声明、现已补上
// ============================================================
//
// 服务端删图片**只认** delete_image_ids,而请求里"这张图不在 image_list
// 里"**不会**触发删除。所以少了这个字段时,前端编辑页"删掉一张图"会
// 表现为:点了删除、保存成功、图片还在。
//
// 前端的 TS 类型本来就有它(types/product.ts 的 delete_image_ids?: number[]),
// 但 go-zero 会**静默丢弃**未在 .api 声明的字段 —— 现已补进
// UpdateProductFullReq 并在此透传。
//
// **与 SKU 的规则不对称,这是最容易误判的地方**:
//
//	SKU  : 请求里没带的会被软删(以请求为准的全量覆盖)
//	图片 : 只有出现在 delete_image_ids 里的才删(没带就不动)
//
// 所以"提交了不完整的 sku_list"会静默删 SKU,而图片不会。
//
// ============================================================
// spu_status 这条路改不了上下架状态(不是缺口)
// ============================================================
//
// proto 的这条 RPC 本身就没有状态字段 —— 状态只能走 publish / withdraw。
// 前端编辑页会在 payload 里带 spu_status,它在**单体时代也被丢掉**
// (单体这条请求结构体同样没有该字段),故不是回归。
// 但值得知道:**整体更新改不了上下架状态**,要让商品上架得调 publish。
//
// ============================================================
// 空值语义的代价:清空 / 归零做不到
// ============================================================
//
// "空串即不改"意味着 brand / description / main_image 改不成空串,
// priority 改不成 0,category_id 也传不了 0 —— 与前端的表单行为一致
// (表单总是带当前值),但若将来要做"清空品牌"这种操作,必须回到 proto
// 加 optional/指针,而不是在 BFF 里塞空串(那会被服务端当"不改")。
//
// ============================================================
// sku_list 传空数组 = 不动 SKU,而不是"清空 SKU"
// ============================================================
//
// 服务端按 len(sku_list) > 0 判断要不要处理 SKU。故请求里给一个空数组
// 得到的语义是"SKU 保持原样"。它**不会**回 ErrSkuListEmpty ——
// 那个分支在 proto 这条路径上其实不可达(converter 只在 len > 0 时
// 才把列表设进请求结构体)。
//
// 想清空 SKU 目前做不到,也没必要:没有 SKU 的商品上不了架(上架校验
// 要求至少一个 active SKU)。
func (l *UpdateProductFullLogic) UpdateProductFull(req *types.UpdateProductFullReq) (*types.Empty, error) {
	specTemplate, err := jsonText(req.SpecTemplate)
	if err != nil {
		return nil, err
	}

	skuList := make([]*v1_productv1.Sku, 0, len(req.SkuList))
	for _, it := range req.SkuList {
		// 带不带 sku_id 决定"改"还是"增",见上方规则 —— 这里原样透传。
		sku, err := toProtoSkuFromAdminSku(it)
		if err != nil {
			return nil, err
		}
		skuList = append(skuList, sku)
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.UpdateProductFull(ctx, &v1_productv1.UpdateProductFullReq{
		SpuId:        req.Id,
		SpuName:      req.SpuName,
		CategoryId:   optionalInt64(req.CategoryId),
		Brand:        req.Brand,
		Description:  req.Description,
		MainImage:    req.MainImage,
		SpecTemplate: specTemplate,
		Priority:     optionalInt64(req.Priority),
		SkuList:      skuList,
		ImageList:    toProtoImages(req.ImageList),
		// DeleteImageIds 透传。
		//
		// 服务端删图**只认这个字段** —— 不传它,前端"删掉一张图"的操作
		// 到不了服务端(表现为:点了删除、保存成功、图片还在)。
		//
		// **注意图片与 SKU 的删除规则不对称**:
		//	SKU  : 请求里没带的会被软删(以请求为准的全量覆盖)
		//	图片 : 只有出现在这里才删(没带就不动)
		// 所以提交不完整的 sku_list 会静默删 SKU,而图片不会。
		DeleteImageIds: req.DeleteImageIds,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品不存在 / 已上架不许改规格模板 / 规格与模板不匹配 /
		// 目标类目不可用 / SKU 编码重复 / 价格非法 → 400
		return nil, err
	}

	// 服务端的 UpdateProductFullResp 带着合并后的完整详情(注释:"调用方
	// 无需再查一次"),而 .api 这条路由是 returns (Empty) —— 与
	// UpdateProduct 同一处契约收紧,理由见那边的说明。
	return &types.Empty{}, nil
}
