// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListUsers 分页查询用户(管理端)。
//
// ============================================================
// 这一条的响应形状与其它列表接口**都不同**
// ============================================================
//
// 单体 GetUserList 的响应是 handler 内联拼的 gin.H:
//
//	utils.Success(c, gin.H{
//		"list":     users,
//		"total":    total,
//		"page":     page,
//		"pageSize": pageSize,     // ← camelCase
//	})
//
// 而其它列表接口(RBAC 那些)是 proto 直传,键名是 **items / total**。
// 那个 pageSize 是 camelCase 的原因在查询参数那一侧:单体用的是
// c.DefaultQuery("pageSize", "10")(handler/user_handler.go:183),
// 响应里回写的也是同一个名字。
//
// **这是既有的不一致,迁移时不能统一** —— 前端按各自的键解。
// 统一成 items/page_size 会让用户列表页整片空数据。
//
// ============================================================
// 分页参数的默认值兜底
// ============================================================
//
// 单体: page 默认 "1",pageSize 默认 "10"。
//
// .api 里 types.ListUsersReq 的两个字段都是 optional int,
// 不传时为 0。**不能直接把 0 传给下游** —— proto 的 page/page_size
// 是 int32,传 0 时 user-service 要么返回全部要么返回空,
// 两种都不是单体行为。故这里兜底成 1 / 10。
//
// 且**把兜底后的值回写到响应** —— 单体回的也是 DefaultQuery 取到的
// 值(不是请求里的原始值)。客户端不传时响应里会看到
// page=1、pageSize=10,前端据此渲染分页控件。
func (l *ListUsersLogic) ListUsers(req *types.ListUsersReq) (*types.ListUsersResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.ListUsers(ctx, &v1_userv1.ListUsersReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   req.Status,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// items → list(键名转换,见上方说明)。
	//
	// 用 make(..., 0, len) 而不是 var list []types.UserItem:
	// 后者在无数据时是 nil,JSON 序列化成 null —— 而前端有
	// `list.length` 这类写法,null 会抛错。空切片序列化成 []。
	items := resp.GetItems()
	list := make([]types.UserItem, 0, len(items))
	for _, u := range items {
		list = append(list, toUserItem(u))
	}

	return &types.ListUsersResp{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
