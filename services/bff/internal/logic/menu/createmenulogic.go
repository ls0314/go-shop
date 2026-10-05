// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateMenu 创建菜单(管理端)。
//
// ============================================================
// MetaInfo 在这里是**可以直接传的**(与 UpdateMenu 不同)
// ============================================================
//
// 这条走的是 CreateMenuReq{ Menu menu = 1 } —— 传的是完整的
// Menu 消息,里面 MetaInfo 就是 google.protobuf.Struct,
// 类型天然匹配,不需要任何编码约定。
//
// 而 UpdateMenu 走 FieldUpdate(oneof 只有 string/int64/bool),
// 要塞一个对象只能序列化成 JSON 字符串,那个约定我没核实 ——
// 故 **UpdateMenu 不支持改 meta_info**。见 helpers.go 的说明。
//
// 这个不对称是真实的注意点:创建时能设 MetaInfo,更新时改不了。
// 前端若有"编辑菜单"页要改图标/标题(通常放在 meta_info 里),
// 目前只能用其它字段或另开接口。
//
// ============================================================
// 需要把 interface{} 转回 map 才能赋给 proto 的 Struct
// ============================================================
//
// .api 里 MetaInfo 是 interface{}(goctl 不认 any),而 proto 的
// MetaInfo 是 *structpb.Struct。故要用 structpb.NewStruct 转换。
//
// **转换可能失败**:MetaInfo 里若有 JSON 不支持的值(如函数、
// channel、循环引用)会返回 error。那时回 400 而不是静默丢字段 ——
// 菜单的 meta_info 丢了会让前端渲染不出标题/图标,很难排查。
func (l *CreateMenuLogic) CreateMenu(req *types.CreateMenuReq) (*types.CreateMenuResp, error) {
	metaInfo, err := toProtoStruct(req.MetaInfo)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.CreateMenu(ctx, &v1_userv1.CreateMenuReq{
		Menu: &v1_userv1.Menu{
			ParentId:  req.ParentId,
			MenuName:  req.MenuName,
			MenuType:  req.MenuType,
			Icon:      req.Icon,
			RoutePath: req.RoutePath,
			Component: req.Component,
			IsVisible: req.IsVisible,
			IsCache:   req.IsCache,
			SortOrder: req.SortOrder,
			MetaInfo:  metaInfo,
			// MenuId / CreatedAt / Children 不传 ——
			// ID 由服务端生成,时间是 DB 填,Children 只是树的读侧结构
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 路由路径重复 / 父菜单不存在 → 400
		return nil, err
	}

	// 响应是**裸菜单对象**(单体: utils.Success(c, created))
	return &types.CreateMenuResp{
		MenuItem: converter.MenuItem(resp.GetMenu()),
	}, nil
}
