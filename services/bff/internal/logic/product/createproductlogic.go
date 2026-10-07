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

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateProduct 建品:SPU + SKU 列表 + 图片列表,**服务端同一事务**。
//
// ============================================================
// 入参形状差一层(见 convert.go 的 toProtoCreateSpu)
// ============================================================
//
// HTTP 是扁平的(spu_name / sku_list / image_list 平铺),proto 是
// `CreateProductReq{ spu }` —— 一个 Spu 实体包住全部字段。组装在
// convert.go 里做,这里只管调用与错误分类。
//
// ============================================================
// 响应只有 spu_id
// ============================================================
//
// proto 的 CreateProductResp 是 {spu_id, error_msg} —— error_msg 是
// **错误通道**(服务端业务失败时填它并回 nil error),不是响应字段,
// 故这里只回 spu_id,与单体 utils.Success(c, gin.H{"spu_id": spuId}) 一致。
//
// 前端建品后只拿这个 id 做跳转/提示,不依赖更多字段(它要看详情会再查一次)。
//
// ============================================================
// 业务失败**不在这里预检**,全部由服务端判定
// ============================================================
//
// 服务端在建品时会校验:类目存在且是启用的叶子、SKU 列表非空、
// 规格模板与 SKU 的规格组合逐项匹配(含笛卡尔积数量)、SKU 编码唯一、
// 价格 > 0。这些都要读 product_db,BFF 预检等于把同一套规则抄第二遍,
// 必然漂移。服务端返回的 error_msg 会经 Classify 落到 400,文案直接给前端。
//
// **注意一处结果是 503 而不是 400 的情况**:spu_name 为空串会撞 DB 的
// ck_spu_name_not_empty 约束,那是 INSERT 报错 → gRPC error → 基础设施
// 失败。前端的建品表单已在前端校验非空,故正常路径走不到;若端到端
// 测到"空名称回 503",那不是 BFF 的错,而是"客户端的错被算成了服务端故障"。
func (l *CreateProductLogic) CreateProduct(req *types.CreateProductReq) (*types.CreateProductResp, error) {
	spu, err := toProtoCreateSpu(req)
	if err != nil {
		// 只有 spec_values / spec_template 无法序列化时才会走到 ——
		// 经 httpx.Parse 解析出来的值一定是 JSON 可表达的,属"不该发生"。
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.CreateProduct(ctx, &v1_productv1.CreateProductReq{
		Spu: spu,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 类目不可用 / SKU 列表为空 / 规格与模板不匹配 / SKU 编码重复 /
		// 价格非法 → 400,文案由服务端给。
		return nil, err
	}

	return &types.CreateProductResp{
		SpuId: resp.GetSpuId(),
	}, nil
}
