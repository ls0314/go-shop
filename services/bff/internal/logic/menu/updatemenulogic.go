// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateMenu 局部更新菜单(管理端)。
//
// ============================================================
// meta_info 在这条接口上改不了 —— **proto 的设计限制,不是 BFF 的缺口**
// ============================================================
//
// 已核实:整条链路上的三处转换**都只支持三种标量**。
//
//	① 单体 userclient.toProtoFieldValue
//	     string / bool / float64(转 int64)
//	② BFF(本文件)
//	     *string / *bool / *int64
//	③ user-service converter.FieldUpdatesToMap
//	     FieldValue_StringValue / _Int64Value / _BoolValue
//	     (switch 无 default —— 其它类型静默丢弃)
//
// 根因在 proto:FieldValue 是个 oneof,**只有这三种标量**。
// 而 Menu.meta_info 是 datatypes.JSONMap(model 侧)/ Struct(proto 侧)
// —— 一个对象。**没有任何办法把对象塞进 FieldValue**。
//
// 所以**单体同样改不了 meta_info**,这不是 BFF 引入的回归。
// 既有前端若从不尝试改它,说明它本来就没这个能力。
//
// ============================================================
// 由此得出的结论:不要在这里"想办法传过去"
// ============================================================
//
// 两条看似可行的路,都不可行:
//
//	序列化成 JSON 字符串放进 string_value
//	  → 服务端 FieldUpdatesToMap 得到的是 string,
//	    而 mapstructure 要把它填进 JSONMap 字段 —— **类型不匹配**,
//	    会报 decode 错误或填进一个字符串。除非同时改服务端加解析。
//	把对象拆成多个 field
//	  → 需要服务端约定 "meta_info.title" 这种扁平键名,
//	    而 mapstructure 不认识。同样要改服务端。
//
// 要真正支持,得**改 proto**(给 oneof 加一个 Struct/Map 成员)
// 或**改服务端的 FieldUpdatesToMap**(加 string→JSON 的解析)。
// 两者都是跨服务的契约变更,不该在 BFF 这一层悄悄绕过去。
//
// ============================================================
// 因此:req.MetaInfo 被静默忽略
// ============================================================
//
// 后果:CreateMenu 能设 meta_info(走完整 Menu 消息,类型天然匹配),
// UpdateMenu 改不了(走 FieldValue,类型不匹配)。
// 这个不对称与单体一致。
//
// **可见症状**:前端若"只改 meta_info",会收到 400 `请求参数错误`
// (因为 updates 为空)。那不是 bug,是这条路径本就不支持。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// is_visible / is_cache 是指针,这一点在菜单域尤其重要:
// 客户端只想改 menu_name 时,若这两个字段是非指针 bool,
// 它们会以零值 false 一起提交,侧边栏上那个菜单就消失了。
func (l *UpdateMenuLogic) UpdateMenu(req *types.UpdateMenuReq) (*types.UpdateMenuResp, error) {
	updates := compactFields(
		strField("menu_name", req.MenuName),
		strField("icon", req.Icon),
		strField("route_path", req.RoutePath),
		strField("component", req.Component),
		boolField("is_visible", req.IsVisible),
		boolField("is_cache", req.IsCache),
		int64Field("sort_order", req.SortOrder),
		// meta_info 不发送 —— proto 的 FieldValue 装不下对象,见上方说明
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.UpdateMenu(ctx, &v1_userv1.UpdateMenuReq{
		MenuId:  req.Id,
		Updates: updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// proto 注释:UpdateMenuResp.menu 返回**合并后的完整对象**
	return &types.UpdateMenuResp{
		MenuItem: converter.MenuItem(resp.GetMenu()),
	}, nil
}
