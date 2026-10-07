// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package category

import (
	"context"

	v1_productv1 "demo-shop/api/gen/product/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateCategory 局部更新类目。
//
// ============================================================
// 六个可选字段拼成一个 JSON 文本;字段用**指针**表达"传没传"
// ============================================================
//
// proto 的 UpdateCategoryReq 只有 {category_id, updates_json},HTTP 侧是
// {category_name, sort_order, icon_url, is_visible, status, parent_id}。
// 拼装规则写在 convert.go;键名必须是服务端实体的 json tag(服务端
// mapstructure 的 TagName 就是 "json"),不是 proto 字段名、也不是 gorm 列名。
//
// **presence 由指针承担**:.api 里这六个字段全是 `*T`(见 bff.api 的
// UpdateCategoryReq 注释),所以:
//
//	nil      → 前端没传 → 不进 JSON → 服务端不改这个字段
//	非 nil   → 前端传了(哪怕值是零值)→ 进 JSON → 服务端按值改
//
// 这消掉了两处本来无解的缺口:
//
//	is_visible=false 能传出去 → **能隐藏类目**
//	parent_id=0      能传出去 → **能把类目移回根层级**
//
// 非指针写法下这两个操作都表达不出来:false 与 0 都是"没传"的哨兵。
//
// ============================================================
// 单体为什么不需要指针
// ============================================================
//
// 它把 body 绑成 map[string]interface{} 后整体转发
// (category_handler.go:232),map 天然保留"键是否出现"。goctl 生成物
// 做不到 —— 生成的结构体把"没传"和"零值"抹成同一个值,故只能靠指针
// 在绑定层把信息保住。menu 的 is_visible/is_cache、dept 的
// leader_id/sort_order、role 的 is_default 都是同样的先例。
//
// ============================================================
// 两处刻意不做的事
// ============================================================
//
// category_id 不会进 updates_json:标识只认路径参数 :id。(单体 body 是
// map,原则上能改到主键 —— BFF 更严。)
//
// 改父节点时服务端会重推 category_path 并**同步改写所有后代的 path**,
// 也会校验"新父节点不能是自己或自己的后代"(CategoryParentInvalid)。
// 这些 BFF 都不预检:一次 RPC 就够,权威判定在服务端。
//
// 响应是**合并后的完整对象**(proto 的 UpdateCategoryResp 直接带
// Category),故不需要再查一次 —— 前端拿到的就是最新状态。
func (l *UpdateCategoryLogic) UpdateCategory(req *types.UpdateCategoryReq) (*types.Category, error) {
	updatesJSON, err := encodeUpdates(compactFields(
		strPtrField("category_name", req.CategoryName),
		intPtrField("sort_order", req.SortOrder),
		strPtrField("icon_url", req.IconUrl),
		boolPtrField("is_visible", req.IsVisible),
		strPtrField("status", req.Status),
		intPtrField("parent_id", req.ParentId),
	))
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.CategoryRPC.UpdateCategory(ctx, &v1_productv1.UpdateCategoryReq{
		CategoryId:  req.Id,
		UpdatesJson: updatesJSON,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 类目不存在 / 新父节点不存在 / 新父节点是自己或自己的后代 → 400
		return nil, err
	}

	return toCategory(resp.GetCategory()), nil
}
