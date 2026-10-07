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

type UserListProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListProductsLogic {
	return &UserListProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserListProducts 用户端商品列表(**只返回已上架**)。
//
// ============================================================
// 请求参数里有一个字段**没有去处**
// ============================================================
//
// .api 的 UserProductListReq 与管理员那个**共用同一份形状**(前端也
// 确实共用一个 SpuQueryReq),故它带着 spu_status;而 proto 的
// UserGetProductListReq **没有**这个字段(注释:"仅返回已上架")。
//
// 这不是遗漏,是两侧的设计不同:
//
//	服务端:在 logic 里把状态硬编码成 published(DB 路径设 req.SpuStatus,
//	        ES 路径在查询体里写死 term spu_status=published)
//	BFF:   收到了 spu_status 也只能丢掉 —— 没有字段可以放
//
// 故客户端传 spu_status=withdrawn 想"偷看下架商品"是**无效**的
// (被静默忽略,仍只回已上架)。这是正确行为,不需要在 BFF 报错拒绝:
// 多一个参数不影响任何结果,而报错会让"前端复用同一个查询构造器"
// 这件事莫名失败。
//
// ============================================================
// 列表项的字段集比管理端少
// ============================================================
//
// UserSpuListItem 没有 category_id / spu_status / priority / created_at /
// updated_at / total_stock —— 用户端不需要(前端 UserSpuListItem 类型
// 里也确实没有这些)。故映射用 toUserSpuListItems,不要图省事复用
// 管理端那份。
//
// ============================================================
// 分页与 items → list:同管理端
// ============================================================
//
// 请求 page_size / 响应 pageSize(两侧命名不一致,照实实现);proto 的
// items → HTTP 的 list;兜底与封顶(用户端 50)在服务端做。
//
// ============================================================
// 一处服务端行为差异,端到端验证时会看到
// ============================================================
//
// 带 spu_name 关键词且 ES 可用时走 ES 搜索,那条路径**不填 stock**
// (因此列表项里 stock = 0);不带关键词或 ES 降级时走 DB,stock = 该
// SPU 的总库存。同一个商品在"分类浏览"和"关键词搜索"下 stock 不同 ——
// 这是服务端的既知差异(见 services/product 的 UserGetProductList),
// BFF 只做透传,不做补偿(补偿要额外查一次库存,而那会掩盖问题)。
func (l *UserListProductsLogic) UserListProducts(req *types.UserProductListReq) (*types.UserProductListResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.ProductRPC.UserGetProductList(ctx, &v1_productv1.UserGetProductListReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		SpuName:  req.SpuName,
		// 零值 → nil(不过滤),同管理端
		CategoryId: optionalInt64(req.CategoryId),
		Brand:      req.Brand,
		Sort:       req.Sort,
		// **req.SpuStatus 刻意不传**:proto 没有这个字段,见上方说明。
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 商品引用了取不到名称的类目 → 400
		return nil, err
	}

	return &types.UserProductListResp{
		List:     toUserSpuListItems(resp.GetItems()),
		Total:    resp.GetTotal(),
		Page:     int(resp.GetPage()),
		PageSize: int(resp.GetPageSize()),
	}, nil
}
